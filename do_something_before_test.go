package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"strings"
	"testing"
)

func TestRabidaImplDoSomethingBefore_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	var prePaginators []EventSelector
	one := EventSelector{
		Type: ClickEvent,
		Selector: CssSelector{
			Css: "#react > div > div > div.center-content.clearfix > div.left-content > div > div.comment-title-bar > div > span:nth-child(2)",
		},
	}
	prePaginators = append(prePaginators, one)
	job := Job{
		Link:        "https://www.meituan.com/zhoubianyou/1161635/",
		PrePaginate: prePaginators,
		CssSelector: CssSelector{
			Scope: "div.comment-main > div.comment-item",
			Attrs: map[string]CssSelector{
				"date": {
					Css: "a.comment-date",
				},
				"content": {
					Css: "div.user-comment span",
					Before: []EventSelector{
						{
							Type: ClickEvent,
							Condition: Condition{
								Value: "阅读全文",
								CheckFunc: func(text, value string) bool {
									return strings.Contains(text, value)
								},
								ExecSelector: ExecSelector{
									Type: TextEvent,
									Selector: CssSelector{
										Css: "div.user-comment span",
									},
								},
							},
							Selector: CssSelector{
								Css: "div.user-comment span",
							},
						},
					},
				},
			},
		},
		Paginator: CssSelector{
			Css: "#react > div > div > div.center-content.clearfix > div.left-content > div > div.comment-box.clearfix > nav > ul > li.pagination-item.pagination-item-comment.next-btn.active > a",
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
		t.Error(fmt.Sprintf("%+v", err))
	}
}
