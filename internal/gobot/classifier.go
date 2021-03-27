package gobot

import (
	"context"
	"errors"
	"regexp"

	"github.com/spy16/snowman"

	"github.com/mattikus/gobot/internal/gobot/slack"
)

type kind int

type patterns []struct {
	re *regexp.Regexp
	id string
}

// Classifier implements a simple intent classifier using regular expression
// patterns. Zero value is safe for use.
type Classifier struct {
	slack *slack.Slack
	hear  patterns
	reply patterns
}

// NewClassifier returns a pointer to a Classifier instance.
func NewClassifier(slack *slack.Slack) *Classifier {
	return &Classifier{slack: slack}
}

var unknown = snowman.Intent{ID: snowman.SysIntentUnknown}

// Classify detects whether a message is directed at the bot and returns an appropriate intent, if
// possible. Otherwise, it returns SysIntentUnknown.
func (c *Classifier) Classify(_ context.Context, msg snowman.Msg) (snowman.Intent, error) {
	toBot, ok := msg.Attribs["to_bot"].(bool)
	if !ok {
		return snowman.Intent{}, errors.New("can't get to_bool")
	}
	if toBot {
		intent, err := c.classify(c.reply, msg.Body)
		if err != nil {
			return unknown, err
		}
		if intent.ID != snowman.SysIntentUnknown {
			return intent, nil
		}
	}
	return c.classify(c.hear, msg.Body)
}

// classify iterates through each registered pattern and tries to match the msg body
// with it. If a match is identified, all the named expressions are inserted into the
// intent context and returned. If not match is found, returns SysIntentUnknown.
func (c *Classifier) classify(ps patterns, msg string) (snowman.Intent, error) {
	for _, p := range ps {
		matches := p.re.FindStringSubmatch(msg)
		if len(matches) > 0 {
			in := snowman.Intent{ID: p.id, Ctx: map[string]interface{}{}}
			names := p.re.SubexpNames()
			for i, match := range matches {
				in.Ctx[names[i]] = match
			}
			return in, nil
		}
	}
	return snowman.Intent{ID: snowman.SysIntentUnknown}, nil
}

// Hear registers the pattern and intent ID for message which are just overheard, not directed to
// the bot.
func (c *Classifier) Hear(pattern string, intentID string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	c.hear = append(c.hear, struct {
		re *regexp.Regexp
		id string
	}{re: re, id: intentID})
	return nil
}

// Reply registers the pattern and intent ID for messages which are directed to the bot itself.
func (c *Classifier) Reply(pattern string, intentID string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	c.reply = append(c.reply, struct {
		re *regexp.Regexp
		id string
	}{re: re, id: intentID})
	return nil
}
