// Package slack provides a snowman.UI based on the RTM implementation from
// github.com/spy16/snowman.
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spy16/snowman"
)

// regexTmpl is a string which defines a regex used by stripSelf to match messages addresssed
// directly to us and then remove the bot's name.
const regexTmpl = `(?i)^(?:(?P<self>%v|%v|%v)[:,]?)?\s*(?P<text>.*)$`

func New(token, signingSecret, port string, logger logger) *Slack {
	if logger == nil {
		logger = snowman.NoOpLogger{}
	}
	return &Slack{
		client:        slack.New(token),
		signingSecret: signingSecret,
		port:          port,
		logger:        logger,
	}
}

// Slack implements snowman UI using Slack RTM API.
type Slack struct {
	logger

	ctx           context.Context
	cancel        func()
	self          *slack.Bot
	SelfRegex     *regexp.Regexp
	client        *slack.Client
	port          string
	signingSecret string
}

// Listen starts an HTTP server and starts listening for slack events API. Message events
// are pushed to the returned channel.
func (sl *Slack) Listen(ctx context.Context) (<-chan snowman.Msg, error) {
	sl.ctx, sl.cancel = context.WithCancel(context.Background())
	defer sl.cancel()

	out := make(chan snowman.Msg)
	resp, err := sl.client.AuthTest()
	if err != nil {
		return nil, err
	}
	bot, err := sl.client.GetBotInfo(resp.BotID)
	if err != nil {
		return nil, err
	}
	sl.self = bot
	if err != nil {
		return nil, err
	}
	prefixes := []interface{}{
		AddressUser(sl.self.UserID, ""),
		AddressUser(sl.self.UserID, sl.self.Name),
		sl.self.Name,
	}
	re, err := regexp.Compile(fmt.Sprintf(regexTmpl, prefixes...))
	if err != nil {
		return nil, err
	}
	sl.SelfRegex = re
	go sl.listenForEvents(ctx, out)

	return out, nil
}

func (sl *Slack) Say(ctx context.Context, user snowman.User, msg snowman.Msg) error {
	if smsg, ok := msg.Attribs["slack_msg"].(*slackevents.MessageEvent); ok {
		err := sl.SendMessage(msg.Body, smsg)
		if err != nil {
			sl.Warnf("unable to send reply: %v", err)
		}
	}
	return nil
}

func (sl *Slack) listenForEvents(ctx context.Context, out chan<- snowman.Msg) {
	defer close(out)

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		sl.Infof("health check")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sl.Errorf("unable to read event body: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, sl.signingSecret)
		if err != nil {
			sl.Errorf("unable to craft secrets verifier: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			sl.Errorf("unable to parse secrets: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			sl.Errorf("unable to verify secrets: ", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			sl.Errorf("unable to parse EventsAPI JSON message: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var r *slackevents.ChallengeResponse
			sl.Infof("Received challenge response event")
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent
			sl.Debugf("event: %s [data=%#v]", innerEvent.Type, innerEvent.Data)
			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				w.WriteHeader(http.StatusOK)
				sl.handleMessage(ctx, ev, out)
				return
			default:
				sl.Debugf("ignoring unknown event (type=%v)", reflect.TypeOf(ev))
				w.WriteHeader(http.StatusOK)
				return
			}
		}
	})
	sl.Infof("listening for HTTP slack events on port %v", sl.port)
	http.ListenAndServe(":"+sl.port, nil)
}

func (sl *Slack) handleMessage(ctx context.Context, ev *slackevents.MessageEvent, out chan<- snowman.Msg) {
	user, err := sl.client.GetUserInfo(ev.User)
	if err != nil {
		sl.Errorf("GetUserInfo(%q): %v", ev.User, err)
		return
	}

	if ev.User == sl.self.UserID {
		return
	}

	// Determine if the message was directly intended for us, stripping any mentions from the message
	// text.
	tagged := sl.stripSelf(ev)

	snowMsg := snowman.Msg{
		From: snowman.User{
			ID:   user.ID,
			Name: user.RealName,
		},
		Body: ev.Text,
		Attribs: map[string]interface{}{
			"slack_msg":  ev,
			"slack_user": *user,
			"to_bot":     ev.ChannelType == "im" || tagged,
		},
	}

	select {
	case <-ctx.Done():
		return
	case out <- snowMsg:
	}
}

func (sl *Slack) stripSelf(ev *slackevents.MessageEvent) bool {
	if sl.SelfRegex == nil {
		sl.Errorf("unable to find self regex to match with")
		return false
	}

	matches := sl.SelfRegex.FindStringSubmatch(ev.Text)
	if matches == nil || len(matches) < 3 {
		return false
	}

	ev.Text = matches[2]
	return matches[1] != ""
}

// SendMessage sends the text as message to the given channel on behalf of
// the bot instance.
func (sl *Slack) SendMessage(text string, responseTo *slackevents.MessageEvent) error {
	opts := []slack.MsgOption{
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
	}

	if responseTo.ThreadTimeStamp != "" {
		opts = append(opts, slack.MsgOptionTS(responseTo.ThreadTimeStamp))
	}
	_, _, err := sl.client.PostMessage(responseTo.Channel, opts...)
	return err
}

// Self returns details about the currently connected bot user.
func (sl *Slack) Self() *slack.Bot { return sl.self }

// Client returns the underlying Slack client instance.
func (sl *Slack) Client() *slack.Client { return sl.client }

// AddressUser creates the escape sequence for marking a user in a message.
func AddressUser(userID string, userName string) string {
	if userName != "" {
		return fmt.Sprintf("<@%s|%s>:", userID, userName)
	}

	return fmt.Sprintf("<@%s>", userID)
}

type logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type noOpLogger struct{}

func (n noOpLogger) Debugf(string, ...interface{}) {}
func (n noOpLogger) Infof(string, ...interface{})  {}
func (n noOpLogger) Warnf(string, ...interface{})  {}
func (n noOpLogger) Errorf(string, ...interface{}) {}
