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
	// Iframe if true, we will look for the element(s) within the first iframe in the page
	Iframe bool
	// XpathScope Note: only choose one between xpath and css selector
	XpathScope string
	// Xpath xpath expression
	// eg: //*[@id="zz"]/div[2]/ul/li[1]/text()
	// eg: //div[@id="indexCarousel"]//div[@class="item"]//img/@src
	Xpath string
}

type Job struct {
	// Link the url you want to crawl
	Link string
	// CssSelector root css selector
	CssSelector CssSelector
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
		options ...chromedp.ExecAllocatorOption,
	) error

	CrawlWithListeners(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
		confPtr *config.RabiConfig,
		options []chromedp.ExecAllocatorOption,
		listeners ...func(ev interface{}),
	) error
}
