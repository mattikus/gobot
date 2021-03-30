package modules

import (
	"context"
	_ "embed" // For embedding card data.
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
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
	block := slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, msg, false, false), nil, nil)
	return NewMsg(intent.Msg, msg, block), nil
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
	cards := cardData.whiteCard(count)
	var blocks []slack.Block
	for idx, c := range cards {
		msg := ""
		if len(cards) > 1 {
			msg += fmt.Sprintf("*%v.* ", idx+1)
		}
		msg += c
		block := slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, msg, false, false), nil, nil)
		blocks = append(blocks, block)
	}

	return NewMsg(intent.Msg, strings.Join(cards, "\n"), blocks...), nil
}

func init() {
	if err := json.Unmarshal(rawJSON, &cardData); err != nil {
		panic(fmt.Errorf("error unmarshalling JSON cards data: %w", err))
	}

	Reply(`q(?:uestion)? card(?: me)?`, "cards.black", fetchBlack)
	Reply(`card(?: me)? (?P<count>\d*)?`, "cards.white", fetchWhite)
}
