package service

import (
	"context"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
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
	// Iframe if true, we will look for the element(s) within the first iframe in the page. if IframeSelector exist, will look for this.
	Iframe bool
	// IframeSelector specify the iframe selector if have multiple iframe elements
	IframeSelector *CssSelector
	// XpathScope Note: only choose one between xpath and css selector
	XpathScope string
	// Xpath xpath expression
	// eg: //*[@id="zz"]/div[2]/ul/li[1]/text()
	// eg: //div[@id="indexCarousel"]//div[@class="item"]//img/@src
	Xpath    string
	SetAttrs []SetAttribute
	// Before dosomething before retrieve value
	Before    []EventSelector
	Condition *Condition
}

type Job struct {
	// Link the url you want to crawl
	Link string
	// CssSelector root css selector
	CssSelector CssSelector
	// PrePaginate do something before paginate
	PrePaginate []EventSelector
	// Paginator css selector for next page
	Paginator     CssSelector
	PaginatorFunc func(currentPageNo int) CssSelector
	// Limit limits how many pages should be crawled
	Limit         int
	StartPageBtn  CssSelector
	StartPageUrl  string
	EnableCookies HttpCookies
}

type EventSelector struct {
	Type      Event
	Condition Condition
	Selector  CssSelector
}

type HttpCookies struct {
	RawCookies string
	Domain     string
	// Expires hour, default 1 year
	Expires int
}

type SetAttribute struct {
	AttributeName  string
	AttributeValue string
}

type Condition struct {
	Value        string
	CheckFunc    func(text, value string) bool
	ExecSelector ExecSelector
}

type ExecSelector struct {
	Type     Event
	Selector CssSelector
}

type Event string

const (
	ClickEvent              Event = "click"
	SetAttributesValueEvent Event = "setAttributesValue"
	TextEvent               Event = "getTextValue"
	GetAttributeValueEvent  Event = "getAttributeValue"
)

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

	DownloadFile(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(file string),
		confPtr *config.RabiConfig,
		options ...chromedp.ExecAllocatorOption,
	) error
}
