package gobot

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/spy16/snowman"
	snowslack "github.com/spy16/snowman/slack"
)

// Classifier is a type which wraps the snowman.RegexClassifier and extends it for use within Slack,
// adding features to understand how to reply to users. The main reason it exists is to allow
// reading every message and ignoring those which are not directly intended for the bot.
type Classifier struct {
	slack     *snowslack.Slack
	logger    snowman.Logger

	selfRegex *regexp.Regexp
	snowman.RegexClassifier
}

var selfPatternTmpl = `^(?i)(%s|@?%s)[,:\s]*(.*)$`

// MatchSelf takes a message and returns a string containing the input, sans the bot's
// name. It also returns a boolean value indicating whether a match was found.
// 
// On first use, it will compile the necessary regular expression pattern and cache it.
func (c *Classifier) MatchSelf(msg string) (string, bool) {
	// This kinda sucks, but we have to wait until here to initialize the regex because we
	// don't know our bot details until after we've connected.
	if c.selfRegex == nil {
		self := c.slack.Self()
		selfPat := fmt.Sprintf(selfPatternTmpl, snowslack.AddressUser(self.ID, ""), self.Name)
		c.logger.Infof("Self pattern: %q", selfPat)
		c.selfRegex = regexp.MustCompile(selfPat)
	}
	if matches := c.selfRegex.FindStringSubmatch(msg); matches != nil {
		return matches[2], matches[1] != ""
	}
	return msg, false
}

// Classify wraps the standard Classify method from snowman.RegexClassifier and adds in
// functionality related to determining whether a message was sent directly to the bot via
// slack.
func (c *Classifier) Classify(ctx context.Context, msg snowman.Msg) (snowman.Intent, error) {
	var slackMsg slack.Msg
	if smsg, ok := msg.Attribs["slack_msg"]; ok {
		slackMsg = smsg.(slack.Msg)
	}
	if strings.HasPrefix(slackMsg.Channel, "D") {
		return c.RegexClassifier.Classify(ctx, msg)
	}
	if body, ok := c.MatchSelf(msg.Body); ok {
		msg.Body = body
		return c.RegexClassifier.Classify(ctx, msg)
	}
	return snowman.Intent{ID: snowman.SysIntentUnknown}, nil
}

func NewClassifier(slackptr *snowslack.Slack, logger snowman.Logger) *Classifier {
	return &Classifier{
		slack:           slackptr,
		logger:          logger,
		selfRegex:       nil,
		RegexClassifier: snowman.RegexClassifier{},
	}
}
