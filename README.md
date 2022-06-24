[![Go](https://github.com/JohnnyTing/rabida/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/JohnnyTing/rabida/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/JohnnyTing/rabida/branch/master/graph/badge.svg?token=XH87JJTRWS)](https://codecov.io/gh/JohnnyTing/rabida)
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>

### Rabida [中文](https://github.com/JohnnyTing/rabida/blob/master/README_ZH.md)

Rabida is a simply crawler framework based on [chromedp](https://github.com/chromedp/chromedp/) .

### Supported features

- `Pagination`:  specify css selector for next page.
- `PrePaginate`: do something before pagination, such as click button.
- `HttpCookies`: enable browser cookie for current job.
- `Delay And Timeout`:  can customize delay and timeout.
- `AntiDetection`: default loaded anti_detetion script for current job. script sourced
  from [puppeteer-extra-stealth](https://github.com/berstend/puppeteer-extra/tree/master/packages/extract-stealth-evasions#readme)
- `Strict Mode`: useragent、browser、platform must be matched，will be related chrome-mac if true
- `Xpath`: specify xpath expression to lookup elements
- `Iframe`: be able to specify the iframe selector

### Install

```go
go get -u github.com/JohnnyTing/rabida
```

### Configuration

add .env file for your project

```shell
RABI_DELAY=1s,2s
RABI_CONCURRENCY=1
RABI_THROTTLE_NUM=2
RABI_THROTTLE_DURATION=1s
RABI_TIMEOUT=3s
RABI_MODE=headless
RABI_DEBUG=false
RABI_OUT=out
RABI_STRICT=false
RABI_PROXY=
```

### Usage

See [examples](https://github.com/JohnnyTing/rabida/blob/master/examples) for more details



Css Selectior:

```go
func TestRabidaImplCrawl(t *testing.T) {
    conf := config.LoadFromEnv()
    fmt.Printf("%+v\n", conf)
    rabi := NewRabida(conf)
    job := Job{
        Link: "https://tieba.baidu.com/f?kw=nba",
        CssSelector: CssSelector{
            Scope: `#thread_list > li.j_thread_list`,
            Attrs: map[string]CssSelector{
                "title": {
                    Css: "div.threadlist_title > a",
                },
                "date": {
                    Css: "span.threadlist_reply_date",
                },
            },
        },
        Paginator: CssSelector{
            Css: "#frs_list_pager > a.next.pagination-item",
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

Xpath Expression:

```go
func TestRabidaXpathImpl_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := NewRabida(conf)
	job := Job{
		Link: "https://you.ctrip.com/sight/shenzhen26/2778.html",
		CssSelector: CssSelector{
			XpathScope: `//*[@id="commentModule"]/div[@class='commentList']/div`,
			Attrs: map[string]CssSelector{
				"content": {
					Xpath: "//div[@class='commentDetail']",
				},
				"date": {
					Xpath: `//div[@class='commentTime']`,
				},
			},
		},
		Paginator: CssSelector{
			Xpath: "//*[@id='commentModule']//li[@class=' ant-pagination-next' and not(@aria-disabled='true')]",
		},
		Limit: 3,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
		}
		logrus.Printf("currentPageNo: %d\n", currentPageNo)
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, []chromedp.Action{
		chromedp.EmulateViewport(1777, 903, chromedp.EmulateLandscape),
	})
	if err != nil {
		t.Error(fmt.Sprintf("%+v", err))
	}
}
```




