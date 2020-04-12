package main

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/nvulane/news-aggry/pkg/configs"
	"github.com/nvulane/news-aggry/pkg/models"
	"github.com/nvulane/news-aggry/pkg/repositories"
	"github.com/nvulane/news-aggry/services/feedersvc/feeder"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var feedUrls = []string{
	"https://news.ycombinator.com/rss",
	"https://lobste.rs/newest.rss",
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.WithField("ctx", "feeder_main")

	logger.Info("Starting Feeder Service...")
	parser := gofeed.NewParser()

	cfg, err := configs.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("could not parse service configuration")
	}

	repository, err := repositories.NewFeederRepository(&cfg.DBConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("could not initialize repository")
	}
	err = repository.CreateFeed(context.TODO(), models.Feed{
		ID:       0,
		Name:     "Hacker News",
		Created:  time.Now(),
		Updated:  time.Now(),
		Category: "Technology",
		URL:      "https://news.ycombinator.com/rss",
	})
	newsFeeder := feeder.NewFeeder(parser, repository, logger)

	c := cron.New()
	c.AddFunc("@every 1m", func() {
		logger.Info("Updating articles")
		err := newsFeeder.UpdateArticles()
		if err != nil {
			logger.WithError(err).Errorf("failed to update articles")
			return
		}
		logger.Info("Successfully updated articles")
	})
	c.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	logger.WithField("signal", sig).Info("received signal")
	logger.Info("Graceful shutdown successful")
}
