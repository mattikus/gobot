package modules

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/spy16/snowman"
)

var urls = [3]string{
	"http://www.youtube.com/watch?v=gODZzSOelss",
	"http://www.youtube.com/watch?v=rT1nGjGM2p8",
	"https://www.youtube.com/watch?v=mDp-ABzpRX8",
}

//go:embed eastwest.json
var rawPlayerData []byte

type player struct {
	Name    string
	College string
	Image   string
}

type bowl struct {
	East []player
	West []player
}

func (b *bowl) RandEast() player {
	return b.East[rand.Intn(len(b.East))]
}

func (b *bowl) RandWest() player {
	return b.West[rand.Intn(len(b.West))]
}

func (b *bowl) RandPlayer() player {
	n := append(b.East, b.West...)
	return n[rand.Intn(len(n))]
}

var bowlData bowl

func fetchURL(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	n, ok := intent.Ctx["url"].(string)
	if !ok || n == "" {
		n = "1"
	}
	i, err := strconv.Atoi(n)
	if err != nil {
		return snowman.Msg{}, err
	}
	return snowman.Msg{
		Body:    urls[i-1],
		Attribs: intent.Msg.Attribs,
	}, nil
}

func fetchPlayer(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	conference, ok := intent.Ctx["conference"].(string)
	if !ok || conference == "" {
		return snowman.Msg{}, fmt.Errorf("unable to determine conference")
	}
	var p player
	switch strings.ToLower(conference) {
	case "east":
		p = bowlData.RandEast()
	case "west":
		p = bowlData.RandWest()
	default:
		p = bowlData.RandPlayer()
	}

	body := fmt.Sprintf("*Name:* %v\n*College:* %v\n %v", p.Name, p.College, p.Image)
	return NewMsg(intent.Msg, body,
		slack.NewImageBlock(p.Image, p.Name, "", nil),
		slack.NewSectionBlock(nil, []*slack.TextBlockObject{
			slack.NewTextBlockObject(slack.MarkdownType, ">>>*Name:*\n"+p.Name, false, false),
			slack.NewTextBlockObject(slack.MarkdownType, ">>>*College:*\n"+p.College, false, false),
		}, nil)), nil
}

func init() {
	if err := json.Unmarshal(rawPlayerData, &bowlData); err != nil {
		panic(fmt.Errorf("error unmarshalling JSON cards data: %w", err))
	}

	Reply(`(?P<conference>east|west|eastwest)(?: me)?$`, "eastwest.player", fetchPlayer)
	Reply(`eastwest(?: me)? url\s*(?P<url>[123])?$`, "eastwest.url", fetchURL)
}
