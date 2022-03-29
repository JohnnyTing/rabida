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
	// Mode Headless or browser
	Mode string `default:"headless"`
	// Debug if true, it will take full screenshot of every page and output debug logs to stdout
	Debug bool
	// Out screenshot png files output path
	Out string
	// Strict useragent and platform be matched、 platform must not be linux etc..
	Strict bool
	//ProxyServer proxy server
	Proxy string
}

func LoadFromEnv() *RabiConfig {
	var conf RabiConfig
	err := envconfig.Process("rabi", &conf)
	if err != nil {
		logrus.Panicln("Error processing env", err)
	}
	return &conf
}
