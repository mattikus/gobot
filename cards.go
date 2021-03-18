package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/spy16/snowman"
)

//go:embed cah-cards-compact.json
var rawJSON []byte

type BlackCard struct {
	Text string
	Pick int
}

// Cards is a type which holds all the available CAH cards data, split into black and white cards.
type Cards struct {
	White []string
	Black []*BlackCard
}

func (c *Cards) WhiteCard(count int) []string {
	var out []string
	for i := 0; i < count; i++ {
		out = append(out, c.White[rand.Intn(len(c.White))])
	}
	return out
}

func (c *Cards) BlackCard() *BlackCard {
	return c.Black[rand.Intn(len(c.Black))]
}

// cards is a global which holds the embedded cards data after initialization.
var cardData Cards

func init() {
	if err := json.Unmarshal(rawJSON, &cardData); err != nil {
		logger.Fatalf("Error unmarshalling JSON cards data: %v", err)
	}
}

func fetchBlack(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	card := cardData.BlackCard()
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
		Body: strings.Join(cardData.WhiteCard(count), "\n"),
		Attribs: intent.Msg.Attribs,
	}, nil
}

func registerCards(c *classifier, pp *processor) error {
	if err := c.Register(`q(?:uestion)? card(?: me)?`, "cards.black"); err != nil { return err }
	if err := pp.Register("cards.black", fetchBlack); err != nil { return err }

	if err := c.Register(`card(?: me)? (?P<count>\d*)?`, "cards.white"); err != nil { return err }
	if err := pp.Register("cards.white", fetchWhite); err != nil { return err }

	return nil
}
