package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"time"
)

type RabiConfig struct {
	// Delay the next request between from and to milliseconds
	// if only from is specified, delay exactly from milliseconds
	// TODO
	Delay []time.Duration `default:"1s"`
	// Concurrency limits concurrency
	// TODO
	Concurrency int
	// ThrottleNum limits the request rate to ThrottleNum per ThrottleDuration
	// TODO
	ThrottleNum      int           `split_words:"true"`
	ThrottleDuration time.Duration `split_words:"true"`
	// Timeout specify a timeout for each page loading and element lookup
	Timeout time.Duration `default:"10s"`
	// ChromePath specify Chrome browser path in the current system
	ChromePath string
}

func LoadFromEnv() *RabiConfig {
	var conf RabiConfig
	err := envconfig.Process("rabi", &conf)
	if err != nil {
		logrus.Panicln("Error processing env", err)
	}
	return &conf
}
