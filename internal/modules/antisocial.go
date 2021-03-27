package modules

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"text/template"

	"github.com/spy16/snowman"

	"github.com/mattikus/gobot/internal/gobot/slack"
)

var antisocials = map[string]struct {
	withTarget string
	noTarget   string
}{
	"maul": {
		withTarget: "{{.Subject}} mauls {{.Target}} in angry bear-like fashion.",
		noTarget:   "{{.Subject}} RAAAHR!!",
	},
	"grid": {
		withTarget: "{{.Subject}} grids {{.Target}} in angry grid-like fashion.",
		noTarget:   "{{.Subject}} Grid?",
	},
	"charades": {
		withTarget: "{{.Subject}}, with finger on nose, points to {{.Target}}.",
		noTarget:   "{{.Subject}} Sounds like...",
	},
	"greet": {
		withTarget: "{{.Subject}} [to {{.Target}}]: Corned beef monkey monkey monkey butt!",
		noTarget:   "{{.Subject}} FUECKER!",
	},
	"nelson": {
		withTarget: "{{.Subject}} [to {{.Target}}]: HAW HAW!",
		noTarget:   "{{.Subject}} I *said* HAW HAW!",
	},
	"ivan": {
		withTarget: "{{.Subject}} chuckels maelvoelntly at {{.Target}}.",
		noTarget:   "{{.Subject}} He types good.",
	},
	"flame": {
		withTarget: "{{.Subject}} sets {{.Target}} on fire.",
		noTarget:   "{{.Subject}} YOU MORON! HITLER!!",
	},
	"cheese": {
		withTarget: "{{.Subject}} [to {{.Target}}]: I like cheese.",
		noTarget:   "{{.Subject}} Behold the power of cheese!",
	},
	"chuck": {
		withTarget: "{{.Subject}} wishes {{.Target}} a happy birthday.  And then this big hairy mouse with very bored eyes comes in and dances with {{.Target}}.",
		noTarget:   "{{.Subject}} all the time singing music (think chipmonks on speed) broadcast in mono over the sound system with peak levels that make the speakers crackle",
	},
	"fire": {
		withTarget: "{{.Target}}: You're fired.",
		noTarget:   "EVACUATE THE BUILDING!",
	},
	"pound": {
		withTarget: "{{.Subject}} pounds and pounds {{.Target}} with a shovel.",
		noTarget:   "{{.Subject}} I'll take 'Things you just want to pound and pound with a shovel' for $300, Alex.'",
	},
	"eye": {
		withTarget: "{{.Subject}} eyes {{.Target}} warily.",
		noTarget:   "{{.Subject}} nay.",
	},
	"thank": {
		withTarget: "{{.Subject}} [to {{.Target}}]: Thanks {{.Target}}! BOK BOK!",
		noTarget:   "{{.Subject}} I DON'T KNOW WHAT TO SAY WHEN YOU SAY THAT.",
	},
	"back": {
		withTarget: "{{.Subject}} slowly backs away from {{.Target}}, careful not to make eye contact.",
		noTarget:   "{{.Subject}} Little in the middle but ya got much...",
	},
	"peer": {
		withTarget: "{{.Subject}} peers at {{.Target}} suspiciously.",
		noTarget:   "{{.Subject}} peers at nothing in particular for no good reason.",
	},
}

func triggers() []string {
	var keys []string
	for k := range antisocials {
		keys = append(keys, k)
	}
	return keys
}

func triggerExp() string {
	return fmt.Sprintf(`!(?P<trigger>rand|%v)\s*(?P<target>.*)?$`, strings.Join(triggers(), "|"))
}

func randTrigger() string {
	ts := triggers()
	return ts[rand.Intn(len(ts))]
}

func Antisocial(_ context.Context, intent snowman.Intent) (snowman.Msg, error) {
	t := intent.Ctx["trigger"].(string)
	if t == "rand" {
		t = randTrigger()
	}

	trigger, ok := antisocials[t]
	if !ok {
		return snowman.Msg{}, fmt.Errorf("can't find trigger named %q", trigger)
	}

	s := trigger.noTarget
	target, ok := intent.Ctx["target"].(string)
	if ok && target != "" {
		s = trigger.withTarget
	}

	tmpl, err := template.New("").Parse(s)
	if err != nil {
		return snowman.Msg{}, fmt.Errorf("unable to parse template: %w", err)
	}

	body := &strings.Builder{}
	subject := slack.AddressUser(intent.Msg.From.ID, "")
	tmpl.Execute(body, &struct {
		Subject string
		Target  string
	}{subject, target})
	return snowman.Msg{
		Body:    body.String(),
		Attribs: intent.Msg.Attribs,
	}, nil
}

func init() {
	Hear(triggerExp(), "antisocial", Antisocial)
}
