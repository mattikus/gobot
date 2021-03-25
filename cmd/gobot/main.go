package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattikus/gobot/internal/gobot"
	"github.com/mattikus/gobot/internal/gobot/slack"
	"github.com/mattikus/gobot/internal/modules"

	"github.com/sirupsen/logrus"
	"github.com/spy16/snowman"
)

var log *logrus.Logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	},
	Level: logrus.DebugLevel,
	Hooks: make(logrus.LevelHooks),
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	go cancelOnInterrupt(cancel, log)

	name := os.Getenv("BOT_NAME")
	if name == "" {
		name = "gobot"
	}

	token := os.Getenv("API_TOKEN")
	secret := os.Getenv("SIGNING_SECRET")
	port := os.Getenv("PORT")
	slackUI := slack.New(token, secret, port, log)

	rc := &snowman.RegexClassifier{}
	proc := gobot.NewProcessor()

	if err := modules.Register(rc, proc); err != nil {
		log.Fatalf("Error registering modules: %v", err)
	}

	if err := snowman.Run(ctx,
		snowman.WithName(name),
		snowman.WithLogger(log),
		snowman.WithUI(slackUI),
		snowman.WithClassifier(rc),
		snowman.WithProcessor(proc),
	); err != nil {
		log.Fatalf("bot exited with error: %v", err)
	}

}

func cancelOnInterrupt(cancel context.CancelFunc, logger snowman.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infof("terminating (signal: %v)", sig)
	cancel()
}
