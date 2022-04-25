package service

import (
	"context"

	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
)

//go:generate go-doudou name --file $GOFILE

type CssSelector struct {
	Css string `json:"css"`
	// Attr default is innerText
	Attr string `json:"attr"`
	// Scope supply a scope to each selector
	// In jQuery, this would look something like this: $(scope).find(selector)
	Scope string `json:"scope"`
	// Attrs map each attribute to a css selector. when Attrs equals nil, stop recursively populating
	Attrs map[string]CssSelector `json:"attrs"`
	// Iframe if true, we will look for the element(s) within the first iframe in the page. if IframeSelector exist, will look for this.
	Iframe bool `json:"iframe"`
	// IframeSelector specify the iframe selector if have multiple iframe elements
	IframeSelector *CssSelector `json:"iframeSelector"`
	// XpathScope Note: only choose one between xpath and css selector
	XpathScope string `json:"xpathScope"`
	// Xpath xpath expression
	// eg: //*[@id="zz"]/div[2]/ul/li[1]/text()
	// eg: //div[@id="indexCarousel"]//div[@class="item"]//img/@src
	Xpath    string         `json:"xpath"`
	SetAttrs []SetAttribute `json:"setAttrs"`
	// Before dosomething before retrieve value
	Before    []EventSelector `json:"before"`
	Condition *Condition      `json:"condition"`
}

type Job struct {
	// Link the url you want to crawl
	Link string `json:"link"`
	// CssSelector root css selector
	CssSelector CssSelector `json:"cssSelector"`
	// PrePaginate do something before paginate
	PrePaginate []EventSelector `json:"prePaginate"`
	// Paginator css selector for next page
	Paginator     CssSelector `json:"paginator"`
	PaginatorFunc func(currentPageNo int) CssSelector
	// Limit limits how many pages should be crawled
	Limit         int         `json:"limit"`
	StartPageBtn  CssSelector `json:"startPageBtn"`
	StartPageUrl  string      `json:"startPageUrl"`
	EnableCookies HttpCookies `json:"enableCookies"`
}

type EventSelector struct {
	Type      Event       `json:"type"`
	Condition Condition   `json:"condition"`
	Selector  CssSelector `json:"selector"`
}

type HttpCookies struct {
	RawCookies string `json:"rawCookies"`
	Domain     string `json:"domain"`
	// Expires hour, default 1 year
	Expires int `json:"expires"`
}

type SetAttribute struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
}

type Condition struct {
	Value        string `json:"value"`
	CheckFunc    func(text, value string) bool
	ExecSelector ExecSelector `json:"execSelector"`
}

type ExecSelector struct {
	Type     Event       `json:"type"`
	Selector CssSelector `json:"selector"`
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

	CrawlScroll(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ret []interface{}, cursor int, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
	) error

	CrawlScrollWithConfig(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ret []interface{}, cursor int, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
		conf config.RabiConfig,
		options ...chromedp.ExecAllocatorOption,
	) error

	CrawlScrollWithListeners(ctx context.Context, job Job,
		// callback process result
		// abort pagination if it returns true
		callback func(ctx context.Context, ret []interface{}, cursor int, currentPageNo int) bool,
		// actions before navigation
		before []chromedp.Action,
		// actions after navigation
		after []chromedp.Action,
		confPtr *config.RabiConfig,
		options []chromedp.ExecAllocatorOption,
		listeners ...func(ev interface{}),
	) error
}
