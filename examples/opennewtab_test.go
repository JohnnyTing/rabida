package examples

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	service "github.com/JohnnyTing/rabida"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestRabidaImplOpenNewTab_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := service.NewRabida(conf)
	job := service.Job{
		Link: "http://www.shenyang.gov.cn/zwgk/zcwj/zfwj/",
		CssSelector: service.CssSelector{
			Scope: `.list-sp .title_futi_time`,
			Attrs: map[string]service.CssSelector{
				"title": {
					Css: ".title > a",
				},
				"link": {
					Css:  ".title > a",
					Attr: "href",
				},
				"date": {
					Css: ".time_pub",
				},
			},
		},
		Paginator: service.CssSelector{
			Css: ".fanye > a.h12:nth-last-child(4)",
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
