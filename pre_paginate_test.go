package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestRabidaImplPrePaginate_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	var prePaginators []EventSelector
	one := EventSelector{
		Type: SetAttributesValueEvent,
		Selector: CssSelector{
			Css: "#selectSort > ul",
			SetAttrs: []SetAttribute{
				{
					AttributeName:  "style",
					AttributeValue: "display: block;",
				},
			},
		},
	}
	two := EventSelector{
		Type: ClickEvent,
		Selector: CssSelector{
			Css: "#selectSort > ul > li:nth-child(3) > a",
		},
	}

	prePaginators = append(prePaginators, one, two)
	job := Job{
		Link:        "https://you.ctrip.com/food/27/236629.html",
		PrePaginate: prePaginators,
		CssSelector: CssSelector{
			Scope: "#sightcommentbox > div.comment_single",
			Attrs: map[string]CssSelector{
				"date": {
					Css: "ul > li.from_link > span.f_left > span > em",
				},
				"content": {
					Css: "ul > li.main_con > span",
				},
			},
		},
		Paginator: CssSelector{
			Css: "#sightcommentbox > div.ttd_pager.cf > div > a.nextpage:not(.disabled)",
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
