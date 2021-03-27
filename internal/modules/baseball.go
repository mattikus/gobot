package modules

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/spy16/snowman"
)

const playerURLBase = "http://www.mcs.anl.gov/~acherry/bb-images"

var playerNames = []string{
	"Sleve McDichael", "Onson Sweemey", "Darryl Archideld",
	"Anatoli Smorin", "Rey McSriff", "Glenallen Mixon",
	"Mario McRlwain", "Raul Chamgerlain", "Kevin Nogilny",
	"Tony Smehrik", "Bobson Dugnutt", "Willie Dustice",
	"Jeromy Gride", "Scott Dourque", "Shown Furcotte",
	"Dean Wesrey", "Mike Truk", "Dwigt Rortugal",
	"Tim Sandaele", "Karl Dandleton", "Mike Sernandez",
	"Todd Bonzalez", "Wilson Chul Lee", "Nert Bisels",
	"Kenn Nitvarn", "Fergit Hote", "Coll Bitzron",
	"Lood Janglosti", "Taenis Tellron", "Marnel Hary",
	"Dony Olerberz", "Gin Ginlons", "Wob Wonkoz",
	"Tanny Mlitnirt", "Hudgyn Sasdarl", "Fraven Pooth",
	"Rarr Dick", "Dorse Hintline", "Roy Gamo",
	"Tenpe Laob", "Varlin Genmist", "Pott Korhil",
	"Am O'Erson", "Snarry Shitwon", "Bobs Peare",
	"Renly Mlynren", "Ceynei Doober", "Hom Wapko",
	"Odood Jorgeudey", "Gary Banps", "Jaris Forta",
	"Erl Jivlitz", "Lenn Wobses", "Dan Boyo",
	"Yans Loovensan", "Mob Welronz", "Bannoe Rodylar",
	"Al Swermirstz", "Jinneil Robenko", "Bobson Allcock Dugnut",
	"Chicken Nutlugget",
}

func randPlayer() string {
	return playerNames[rand.Intn(len(playerNames))]
}

func randImg() string {
	return fmt.Sprintf("%v/%v.jpg", playerURLBase, rand.Intn(860)+1)
}

func init() {
	Reply(`baseball(?: me)?`, "baseball.player", func(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
		return snowman.Msg{
			Body:    fmt.Sprintf(">*Player*: %v\n%v", randPlayer(), randImg()),
			Attribs: intent.Msg.Attribs,
		}, nil
	})
}
