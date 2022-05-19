package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

//go:generate go-doudou name --file $GOFILE

type RabiConfig struct {
	// Delay the next request between from and to milliseconds
	// if only from is specified, delay exactly from milliseconds
	Delay []time.Duration `default:"2s,3s" json:"delay"`
	// Concurrency limits concurrency
	// TODO
	Concurrency int `json:"concurrency"`
	// ThrottleNum limits the request rate to ThrottleNum per ThrottleDuration
	// TODO
	ThrottleNum      int           `split_words:"true" json:"throttleNum"`
	ThrottleDuration time.Duration `split_words:"true" json:"throttleDuration"`
	// Timeout specify a timeout for each page loading and element lookup
	Timeout time.Duration `default:"10s" json:"timeout"`
	// Mode Headless or browser
	Mode string `default:"headless" json:"mode"`
	// Debug if true, it will take full screenshot of every page and output debug logs to stdout
	Debug bool `json:"debug"`
	// Out screenshot png files output path
	Out string `default:"out" json:"out"`
	// Strict useragent、browser、platform should be matched, platform should not be linux etc..
	Strict bool `json:"strict"`
	//ProxyServer proxy server
	Proxy string `json:"proxy"`
	// ScrollType scroll should be scrollBy and scrollTo. default is scrollBy
	ScrollType string `json:"scroll_type" default:"scrollBy"`
	// ScrollTop refer to window.scrollTo parameters
	ScrollTop int `json:"scroll_top" default:"800"`
	// ScrollLeft refer to window.scrollTo parameters
	ScrollLeft int `json:"scroll_left" default:"0"`
}

func LoadFromEnv() *RabiConfig {
	var conf RabiConfig
	err := envconfig.Process("rabi", &conf)
	if err != nil {
		logrus.Panicln("Error processing env", err)
	}
	return &conf
}
