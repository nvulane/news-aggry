package feeder

import (
	"context"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/nvulane/news-aggry/pkg/models"
	"github.com/nvulane/news-aggry/pkg/repositories"
	"github.com/sirupsen/logrus"
	"sort"
	"sync"
)

type FeederService interface {
	UpdateArticles() error
}

type feederService struct {
	logger *logrus.Logger
	parser *gofeed.Parser
	repo   repositories.FeedRepository
}

func NewFeeder(parser *gofeed.Parser, repo repositories.FeedRepository, logger *logrus.Logger) FeederService {
	return &feederService{
		repo: repo,
		logger: logger.WithField("ctx", "feederService").Logger,
		parser: parser,
	}
}

func (f *feederService) UpdateArticles() error {
	var lock sync.Mutex
	var wg sync.WaitGroup
	var articles []models.Article
	dbFeeds, err := f.repo.GetAllFeeds(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to fetch feed URLs: %w", err)
	}
	for _, feedInfo := range dbFeeds {
		wg.Add(1)
		go func(feedUrl string) {
			defer wg.Done()
			feed, err := f.parser.ParseURL(feedUrl)
			if err != nil {
				f.logger.WithError(err).WithField("url", feedUrl).Error("unable to parse feed")
				return
			}
			for _, item := range feed.Items {
				wg.Add(1)
				go func(item *gofeed.Item) {
					defer wg.Done()
					article := models.Article{
						Title:       item.Title,
						Description: item.Description,
						Link:        item.Link,
					}
					if item.PublishedParsed != nil {
						article.Published = *item.PublishedParsed
					}

					lock.Lock()
					articles = append(articles, article)
					lock.Unlock()

				}(item)
			}
		}(feedInfo.URL)
	}
	wg.Wait()
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Published.Before(articles[j].Published)
	})
	err = f.repo.InsertArticles(context.TODO(), articles)
	if err != nil {
		return fmt.Errorf("unable to insert articles: %w", err)
	}
	return nil
}
