package gobot

import (
	"context"
	_ "embed" // For embedding card data.
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/spy16/snowman"
)

//go:embed cah-cards-compact.json
var rawJSON []byte

// cardData is a global which holds the embedded cards data after initialization.
var cardData cards

type blackCard struct {
	Text string
	Pick int
}

// Cards is a type which holds all the available CAH cards data, split into black and white cards.
type cards struct {
	White []string
	Black []*blackCard
}

func (c *cards) whiteCard(count int) []string {
	var out []string
	for i := 0; i < count; i++ {
		out = append(out, c.White[rand.Intn(len(c.White))])
	}
	return out
}

func (c *cards) blackCard() *blackCard {
	return c.Black[rand.Intn(len(c.Black))]
}

func fetchBlack(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	card := cardData.blackCard()
	msg := card.Text
	if card.Pick > 1 {
		msg = fmt.Sprintf("*(Pick %v)* %v", card.Pick, msg)
	}

	return snowman.Msg{Body: msg, Attribs: intent.Msg.Attribs}, nil
}

func fetchWhite(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	var err error
	count := 1
	if c, ok := intent.Ctx["count"]; ok && c != "" {
		count, err = strconv.Atoi(c.(string))
		if err != nil {
			return snowman.Msg{}, err
		}
	}

	return snowman.Msg{
		Body:    strings.Join(cardData.whiteCard(count), "\n"),
		Attribs: intent.Msg.Attribs,
	}, nil
}

// RegisterCards takes a gobot Classifier and Processor and adds matchers for functionality within
// this module is added to an instance of gobot.
func RegisterCards(c *Classifier, pp *Processor) error {
	if err := json.Unmarshal(rawJSON, &cardData); err != nil {
		return fmt.Errorf("error unmarshalling JSON cards data: %w", err)
	}

	if err := c.Register(`q(?:uestion)? card(?: me)?`, "cards.black"); err != nil { return err }
	if err := pp.Register("cards.black", fetchBlack); err != nil { return err }

	if err := c.Register(`card(?: me)? (?P<count>\d*)?`, "cards.white"); err != nil { return err }
	if err := pp.Register("cards.white", fetchWhite); err != nil { return err }

	return nil
}