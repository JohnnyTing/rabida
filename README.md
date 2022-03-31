### Rabida 

Rabida is simply crawler framework based on [chromedp]() .

### Supported features

- `Pagination`:  specify css selector for next page.
- `PrePaginate`: do something before pagination, such as click button.
- `HttpCookies`: enable browser cookie for current job.
- `Delay And Timeout`:  can customize delay and timeout.
- `AntiDetection`: default loaded anti_detetion script for current job. script sourced from [puppeteer-extra-stealth](https://github.com/berstend/puppeteer-extra/tree/master/packages/extract-stealth-evasions#readme)
- `Strict Mode`: useragent、browser、platform must be matched，will be related chrome-mac if true

### Install

```go
go get -u https://github.com/JohnnyTing/rabida
```

### Usage

See [examples](https://github.com/JohnnyTing/rabida/blob/master/svcimpl_test.go) for more details

```go
func TestRabidaImpl_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://zjt.fujian.gov.cn/xxgk/zxwj/zxwj/",
		CssSelector: CssSelector{
			Scope: `.box>div>div:not([style="display: none;"])>div.gl_news`,
			Attrs: map[string]CssSelector{
				"title": {
					Css: ".gl_news_top_tit",
				},
				"link": {
					Css:  "a",
					Attr: "href",
				},
				"date": {
					Css: ".gl_news_top_rq",
				},
			},
		},
		Paginator: CssSelector{
			Css:  ".c-txt>a:nth-last-of-type(2)",
			Attr: "href",
		},
		Limit: 3,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, []chromedp.Action{
		chromedp.EmulateViewport(1777, 903, chromedp.EmulateLandscape),
	})
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
```

