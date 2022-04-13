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

func TestRabidaImplNextPage_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := service.NewRabida(conf)
	job := service.Job{
		Link: "http://www.xm.gov.cn/zfxxgk/xxgkznml/szfgz/szfwz/",
		CssSelector: service.CssSelector{
			Scope: `.tit1+table tr:not(:first-child)`,
			Attrs: map[string]service.CssSelector{
				"title": {
					Css: ".info-extra .info_tit:last-child>a:last-child,.info-extra .info_tit:last-child>span:last-child",
				},
				"link": {
					Css:  "table a",
					Attr: "href",
				},
				"date": {
					Css: ":scope table td:last-child",
				},
			},
		},
		Paginator: service.CssSelector{
			Css: ".fy_tit_l a.next",
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
