package gobot

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/spy16/snowman"
)

// Processor is a type which implements the snowman.Processor interface. It has a registry of
// actions and understands how to map an intent to an action, registered by a separate module.
type Processor struct {
	actions map[string]snowman.ProcessorFunc
}

// Process implements the Process method for a snowman.Processor interface.
func (pp *Processor) Process(ctx context.Context, intent snowman.Intent) (snowman.Msg, error) {
	slackMsg, ok := intent.Msg.Attribs["slack_msg"].(slack.Msg)
	if !ok {
		return snowman.Msg{}, fmt.Errorf("unable to parse message")
	}
	slackMsg.Timestamp = ""
	intent.Msg.Attribs["slack_msg"] = slackMsg

	if action, ok := pp.actions[intent.ID]; ok {
		return action(ctx, intent)
	}
	return snowman.Msg{}, nil
}

// Register adds a given processor function to the actions registry for and instance of the
// Processor type.
func (pp *Processor) Register(intentID string, fun snowman.ProcessorFunc) error {
	if pp == nil || pp.actions == nil {
		return fmt.Errorf("unable to register")
	}
	if _, found := pp.actions[intentID]; found {
		return fmt.Errorf("action with name already exists")
	}
	pp.actions[intentID] = fun
	return nil
}

// NewProcessor is a constructor which returns a pointer to an instance of Processor.
func NewProcessor() *Processor {
	pp := &Processor{}
	pp.actions = make(map[string]snowman.ProcessorFunc)
	return pp
}

