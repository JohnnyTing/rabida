package examples

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	. "github.com/JohnnyTing/rabida"
	"github.com/JohnnyTing/rabida/config"
	"testing"
)

func TestRabidaCrawlCtrip(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "https://you.ctrip.com/sight/shenzhen26/2778.html",
		CssSelector: CssSelector{
			Scope: `#commentModule > div.commentList > div.commentItem`,
			Attrs: map[string]CssSelector{
				"content": {
					Css: "div.contentInfo > div.commentDetail",
				},
				"date": {
					Css: "div.contentInfo > div.commentFooter > div.commentTime",
				},
			},
		},
		Paginator: CssSelector{
			Css: "#commentModule > div.myPagination > ul > li.ant-pagination-next[aria-disabled='false']",
		},
		Limit: 5,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func TestRabidaCrawlCtripFromLasted(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	var prePaginators []EventSelector
	one := EventSelector{
		Type: ClickEvent,
		Selector: CssSelector{
			Css: "#commentModule > div.sortList > span:nth-child(2)",
		},
	}
	prePaginators = append(prePaginators, one)

	job := Job{
		Link:        "https://you.ctrip.com/sight/shenzhen26/2778.html",
		PrePaginate: prePaginators,
		CssSelector: CssSelector{
			Scope: `#commentModule > div.commentList > div.commentItem`,
			Attrs: map[string]CssSelector{
				"content": {
					Css: "div.contentInfo > div.commentDetail",
				},
				"date": {
					Css: "div.contentInfo > div.commentFooter > div.commentTime",
				},
			},
		},
		Paginator: CssSelector{
			Css: "#commentModule > div.myPagination > ul > li.ant-pagination-next[aria-disabled='false']",
		},
		Limit: 5,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
