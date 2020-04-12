package configs

import (
	"fmt"
	"github.com/jessevdk/go-flags"
)

type Config struct {
	DBConfig
}

type DBConfig struct {
	Host        string `long:"host" env:"HOST" default:"localhost"`
	Port        int    `long:"port" env:"PORT" default:"27017"`
	DBName      string `long:"dbname" env:"DBNAME" default:"news_feed"`
	User        string `long:"user" env:"USER"`
	Password    string `long:"password" env:"PASSWORD"`
	MaxPoolSize int    `long:"maxpoolsize" env:"MAXPOOLSIZE"`
	Timeout     int    `long:"timeout" env:"TIMEOUT" default:"10"`
}

func (c DBConfig) GetDSN() string {
	//dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DBName)
	dsn := fmt.Sprintf("mongodb://%s:%d/%s?sslmode=disable", c.Host, c.Port, c.DBName)
	return dsn
}

func NewConfig() (*Config, error) {
	conf := &Config{}
	parser := flags.NewParser(conf, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return conf, nil
}
