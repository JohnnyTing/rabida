package service

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/unionj-cloud/rabida/config"
)

type CssSelector struct {
	Css string
	// Attr default is innerText
	Attr string
	// Scope supply a scope to each selector
	// In jQuery, this would look something like this: $(scope).find(selector)
	Scope string
	// Attrs map each attribute to a css selector. when Attrs equals nil, stop recursively populating
	Attrs map[string]CssSelector
}

type Job struct {
	// Link the url you want to crawl
	Link string
	// CssSelector root css selector
	CssSelector
	// Paginator css selector for next page
	Paginator CssSelector
	// Limit limits how many pages should be crawled
	Limit        int
	StartPageBtn CssSelector
	StartPageUrl string
}

type Rabida interface {
	Crawl(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
	) error

	CrawlWithConfig(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
		conf config.RabiConfig,
	) error
}
