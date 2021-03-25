// modules is a package which provides the all of the bot functionality.
package modules

import (
	"github.com/mattikus/gobot/internal/gobot"
	"github.com/spy16/snowman"
)

var actions = make(map[string]snowman.ProcessorFunc)
var intents = make(map[string]string)

// Register injects all of the functionality defined within modules.
func Register(c *snowman.RegexClassifier, pp *gobot.Processor) error {
	for regex, intent := range intents {
		if err := c.Register(regex, intent); err != nil {
			return err
		}
	}
	for intent, fun := range actions {
		if err := pp.Register(intent, fun); err != nil {
			return err
		}
	}
	return nil
}
