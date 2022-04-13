package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"testing"
)

func TestRabidaCrawlIframe(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://www.jinan.gov.cn/col/col27544/index.html",
		CssSelector: CssSelector{
			Scope:  "#searchform+table tr",
			Iframe: true,
			IframeSelector: &CssSelector{
				Css: "#zpinfo003",
			},
			Attrs: map[string]CssSelector{
				"title": {
					Css:  "a",
					Attr: "title",
				},
				"date": {
					Css: "td:last-child>span",
				},
				"link": {
					Css:  "a",
					Attr: "href",
				},
			},
		},
		Paginator: CssSelector{
			Css: "a.pgBtn:nth-child(3):not(.disabledTd)",
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
	}, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func TestRabidaCrawlIframe1(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://www.suzhou.gov.cn/szsrmzf/zfxxgkzl/xxgkml.shtml?para=zcwj",
		CssSelector: CssSelector{
			Scope:  `body > form > table > tbody > tr`,
			Iframe: true,
			IframeSelector: &CssSelector{
				Css: "#xxgk_item",
			},
			Attrs: map[string]CssSelector{
				"content": {
					Css:  "a",
					Attr: "title",
				},
				"date": {
					Css: "td:last-child",
				},
				"link": {
					Css:  "a",
					Attr: "href",
				},
			},
		},
		Paginator: CssSelector{
			Css: "//span[@class='upordown']/a[text()='下一页']",
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
	}, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
