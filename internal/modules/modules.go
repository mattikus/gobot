// Package modules is a package which provides the all of the bot functionality.
package modules

import (
	"github.com/mattikus/gobot/internal/gobot"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spy16/snowman"
)

type module struct {
	id    string
	regex string
	fun   snowman.ProcessorFunc
}

var replyModules []module

func Reply(re, id string, fun snowman.ProcessorFunc) {
	replyModules = append(replyModules, module{id, re, fun})
}

var hearModules []module

func Hear(re, id string, fun snowman.ProcessorFunc) {
	hearModules = append(hearModules, module{id, re, fun})
}

// NewMsg takes a message to reply to and creates a new message with the correct room already set to
// reply.
func NewMsg(replyTo snowman.Msg, body string, blocks ...slack.Block) snowman.Msg {
	var channel string
	if omsg, ok := replyTo.Attribs["slack_msg"].(*slackevents.MessageEvent); ok {
		channel = omsg.Channel
	}
	return snowman.Msg{
		Body: body,
		Attribs: map[string]interface{}{
			"slack_channel": channel,
			"slack_blocks":  blocks,
		},
	}
}

// Register injects all of the functionality defined within modules.
func Register(c *gobot.Classifier, pp *gobot.Processor) error {
	for _, i := range hearModules {
		if err := c.Hear(i.regex, i.id); err != nil {
			return err
		}
		if err := pp.Register(i.id, i.fun); err != nil {
			return err
		}
	}
	for _, i := range replyModules {
		if err := c.Reply(i.regex, i.id); err != nil {
			return err
		}
		if err := pp.Register(i.id, i.fun); err != nil {
			return err
		}
	}
	return nil
}
