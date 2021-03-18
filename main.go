package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spy16/snowman"
	snowslack "github.com/spy16/snowman/slack"
)


var logger *logrus.Logger

func main() {
	rand.Seed(time.Now().UnixNano())
	logger = logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	go cancelOnInterrupt(cancel, logger)

	name := os.Getenv("BOT_NAME")
	if name == "" {
		name = "lester"
	}

	token := os.Getenv("API_TOKEN")
	slackUI := snowslack.New(token, logger)

	rc := &classifier{slack: slackUI}
	proc := NewProcessor()

	if err := registerCards(rc, proc); err != nil {
		logger.Fatalf("Error registering cards.go: %v", err)
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
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Infof("terminating (signal: %v)", sig)
	cancel()
}