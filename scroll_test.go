package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"testing"
)

func TestRabidaImplCrawlScroll(t *testing.T) {
	t.Run("CrawlScroll", func(t *testing.T) {
		conf := config.LoadFromEnv()
		fmt.Printf("%+v\n", conf)
		rabi := NewRabida(conf)
		job := Job{
			Link: "http://www.news.cn/energy/index.html",
			CssSelector: CssSelector{
				Scope: `#content-list > div.item`,
				Attrs: map[string]CssSelector{
					"title": {
						Css: ".tit > a",
					},
					"date": {
						Css: ".time",
					},
				},
			},
			Paginator: CssSelector{
				Css: "#list > div.xpage-more-btn.look",
			},
			Limit: 10,
		}
		err := rabi.CrawlScroll(context.Background(), job, func(ret []interface{}, nextCursor int, currentPageNo int) bool {
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
			t.Errorf("%+v", err)
		}

	})
}

func TestRabidaImplCrawlScrollXpath(t *testing.T) {
	t.Run("CrawlScrollXpath", func(t *testing.T) {
		conf := config.LoadFromEnv()
		fmt.Printf("%+v\n", conf)
		rabi := NewRabida(conf)
		job := Job{
			Link: "http://www.news.cn/energy/index.html",
			CssSelector: CssSelector{
				XpathScope: `//*[@id="content-list"]/div`,
				Attrs: map[string]CssSelector{
					"title": {
						Xpath: "//div[@class='tit']",
					},
					"date": {
						Xpath: "//div[@class='time']",
					},
				},
			},
			Paginator: CssSelector{
				Css: "#list > div.xpage-more-btn.look",
			},
			Limit: 10,
		}
		err := rabi.CrawlScroll(context.Background(), job, func(ret []interface{}, nextCursor int, currentPageNo int) bool {
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
			t.Errorf("%+v", err)
		}

	})
}
