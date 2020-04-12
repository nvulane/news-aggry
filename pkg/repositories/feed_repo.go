package repositories

import (
	"context"
	"fmt"
	"github.com/nvulane/news-aggry/pkg/configs"
	"github.com/nvulane/news-aggry/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

type FeedRepository interface {
	CreateFeed(ctx context.Context, feed models.Feed) error
	GetAllFeeds(ctx context.Context) ([]models.Feed, error)
	GetArticles(ctx context.Context) ([]models.Article, error)
	InsertArticles(ctx context.Context, articles []models.Article) error
}

type feedRepository struct {
	client   *mongo.Client
	database *mongo.Database
	feeds    *mongo.Collection
	articles *mongo.Collection
	logger   *logrus.Logger
}

func NewFeederRepository(cfg *configs.DBConfig, logger *logrus.Logger) (FeedRepository, error) {
	opts := options.Client()
	opts.SetMaxPoolSize(uint64(cfg.MaxPoolSize))
	opts.ApplyURI(cfg.GetDSN())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	repoLogger := logger.WithField("ctx", "repositories").Logger

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("error verifying connection: %w", err)
	}
	repoLogger.Info("Connected to the repository")
	database := client.Database("news_feed")
	feedCollection := database.Collection("feeds")
	_, err = feedCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"link": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	articleCollection := database.Collection("articles")
	_, err = articleCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"link": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &feedRepository{
		client:   client,
		database: database,
		feeds:    feedCollection,
		articles: articleCollection,
		logger:   repoLogger,
	}, nil
}

func (r *feedRepository) CreateFeed(ctx context.Context, feed models.Feed) error {
	_, err := r.feeds.InsertOne(ctx, feed)
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}
	return nil
}

func (r *feedRepository) GetArticles(ctx context.Context) ([]models.Article, error) {
	return nil, nil
}

func (r *feedRepository) GetAllFeeds(ctx context.Context) ([]models.Feed, error) {
	var feeds []models.Feed
	cur, err := r.feeds.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve feeds: %w", err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		if err := cur.Err(); err != nil {
			return nil, err
		}

		var feed models.Feed
		err = cur.Decode(&feed)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (r *feedRepository) InsertArticles(ctx context.Context, articles []models.Article) error {
	var a []interface{}
	for _, article := range articles {
		a = append(a, article)
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := r.articles.InsertMany(ctx, a, opts)
	if !isDuplicateErr(err) {
		return fmt.Errorf("unable to insert articles: %w", err)
	}
	return nil
}

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "duplicate key error")
}
