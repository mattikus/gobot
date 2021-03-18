package main

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/spy16/snowman"
)

type processor struct {
	actions map[string]snowman.ProcessorFunc
}

func (pp *processor) Process(ctx context.Context, intent snowman.Intent) (snowman.Msg, error) {
	slackMsg := intent.Msg.Attribs["slack_msg"].(slack.Msg)
	slackMsg.Timestamp = ""
	intent.Msg.Attribs["slack_msg"] = slackMsg

	if action, ok := pp.actions[intent.ID]; ok {
		return action(ctx, intent)
	}
	return snowman.Msg{}, nil
}

func (pp *processor) Register(intentID string, fun snowman.ProcessorFunc) error {
	if pp == nil || pp.actions == nil {
		return fmt.Errorf("unable to register")
	}
	if _, found := pp.actions[intentID]; found {
		return fmt.Errorf("action with name already exists")
	}
	pp.actions[intentID] = fun
	return nil
}

func NewProcessor() *processor {
	pp := &processor{}
	pp.actions = make(map[string]snowman.ProcessorFunc)
	return pp
}

