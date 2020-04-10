package main

import (
	"github.com/mmcdole/gofeed"
	"github.com/nvulane/news-aggry/services/feeder"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var feedUrls = []string{
	"https://news.ycombinator.com/rss",
	"https://lobste.rs/newest.rss",
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	log.Info("Starting Feeder Service...")
	parser := gofeed.NewParser()
	c := cron.New()
	c.AddFunc("@every 30m", func() {
		feeder.Poll(feedUrls, parser)
	})
	c.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.WithField("signal", sig).Info("received signal")
	log.Info("Graceful shutdown successful")
}
