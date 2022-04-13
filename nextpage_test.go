package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestRabidaImplNextPage_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := NewRabida(conf)
	job := Job{
		Link: "http://js.wuxi.gov.cn/zfxxgk/xxgkml/fgwjjjd/bmwj/index.shtml",
		CssSelector: CssSelector{
			Scope: `#doclist>li`,
			Attrs: map[string]CssSelector{
				"title": {
					Css:  "a",
					Attr: "title",
				},
				"link": {
					Css:  "a",
					Attr: "href",
				},
				"date": {
					Css: "span",
				},
			},
		},
		Paginator: CssSelector{
			Css: ".next",
		},
		Limit: 3,
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
	}
	if conf.Mode == "headless" {
		opts = append(opts, chromedp.Headless)
	}

	ctx := context.Background()

	err := rabi.Crawl(ctx, job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
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
