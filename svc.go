package service

import "context"

type CssSelector struct {
	Css string
	// Attr default is innerText
	Attr string
}

type Job struct {
	// Link the url you want to crawl
	Link string
	// Scope supply a scope to each selector
	// In jQuery, this would look something like this: $(scope).find(selector)
	Scope string
	// Attrs map each attribute to a css selector
	Attrs map[string]CssSelector
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
		callback func(ret []map[string]string, nextPageUrl string, currentPageNo int) bool) error
}
