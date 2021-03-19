package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattikus/gobot/internal/gobot"
	"github.com/mattikus/gobot/internal/modules"

	"github.com/sirupsen/logrus"
	"github.com/spy16/snowman"
	snowslack "github.com/spy16/snowman/slack"
)


var logger *logrus.Logger = logrus.New()

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	go cancelOnInterrupt(cancel, logger)

	name := os.Getenv("BOT_NAME")
	if name == "" {
		name = "gobot"
	}

	token := os.Getenv("API_TOKEN")
	slackUI := snowslack.New(token, logger)

	rc := gobot.NewClassifier(slackUI, logger)
	proc := gobot.NewProcessor()

	if err := modules.Register(rc, proc); err != nil {
		logger.Fatalf("Error registering modules: %v", err)
	}

	if err := snowman.Run(ctx,
		snowman.WithName(name),
		snowman.WithLogger(logger),
		snowman.WithUI(slackUI),
		snowman.WithClassifier(rc),
		snowman.WithProcessor(proc),
	); err != nil {
		logger.Fatalf("bot exited with error: %v", err)
	}

}

func cancelOnInterrupt(cancel context.CancelFunc, logger snowman.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infof("terminating (signal: %v)", sig)
	cancel()
}
