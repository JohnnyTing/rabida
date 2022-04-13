package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	_ "github.com/unionj-cloud/go-doudou/framework/http"
	"testing"
)

func TestRabidaImplDynamicNextPageBtn_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := NewRabida(conf)
	job := Job{
		Link: "https://www.sjz.gov.cn/col/1596014942837/index.html",
		CssSelector: CssSelector{
			Scope: `.nr ul li`,
			Attrs: map[string]CssSelector{
				"title": {
					Css:  "a:first-child",
					Attr: "title",
				},
				"link": {
					Css:  "a:first-child",
					Attr: "href",
				},
				"date": {
					Css: "span.date",
				},
			},
		},
		PaginatorFunc: func(currentPageNo int) CssSelector {
			return CssSelector{
				Css: fmt.Sprintf(`.center #MinyooPage>a[title="当前在第%d页"]+a`, currentPageNo),
			}
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
		t.Error(fmt.Sprintf("%+v", err))
	}
}
