package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/chromedp/chromedp"
	_ "github.com/unionj-cloud/go-doudou/svc/config"
	"github.com/unionj-cloud/rabida/config"
	"testing"
)

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
