package feeder

import (
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Article struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
}

func Poll(feedURLS []string, parser *gofeed.Parser) {
	var wg sync.WaitGroup
	for _, url := range feedURLS {
		wg.Add(1)
		go func(feedUrl string) {
			defer wg.Done()
			feed, err := parser.ParseURL(url)
			if err != nil {
				log.WithError(err).WithField("url", url).Error("unable to parse feed")
				return
			}
			for _, item := range feed.Items {
				wg.Add(1)
				go func(item *gofeed.Item) {
					defer wg.Done()
					article := Article{
						Title:       item.Title,
						Description: item.Description,
						Link:        item.Link,
						Published:   time.Time{},
					}
					log.WithField("article", article).Info("article")
				}(item)
			}
		}(url)
	}
	wg.Wait()
}
