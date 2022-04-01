package examples

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	. "github.com/JohnnyTing/rabida"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"testing"
)

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
