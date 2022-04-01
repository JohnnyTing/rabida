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

func TestRabidaImplPrePaginate_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	var prePaginators []EventSelector
	one := EventSelector{
		Type: ClickEvent,
		Selector: CssSelector{
			Css: "#ibu_hotel_review > div > ul.tab-pane > li > div > div.m-module-bg > div.m-reviewFilter > div.m-reviewFilter-selector > div.right > span > span > div > input[type=input]",
		},
	}
	two := EventSelector{
		Type: ClickEvent,
		Selector: CssSelector{
			Css: "#ibu_hotel_tools > div > ul > li:nth-child(2)",
		},
	}
	prePaginators = append(prePaginators, one, two)
	job := Job{
		Link:        "https://hotels.ctrip.com/hotels/detail/?hotelId=635264&checkIn=2021-01-25&checkOut=2021-01-26&cityId=30&minprice=&mincurr=&adult=1&children=0&ages=&crn=1&curr=&fgt=&stand=&stdcode=&hpaopts=&mproom=&ouid=&shoppingid=&roomkey=&highprice=-1&lowprice=0&showtotalamt=&hotelUniqueKey=#ibu_hotel_review",
		PrePaginate: prePaginators,
		CssSelector: CssSelector{
			Scope: "#ibu_hotel_review > div > ul.tab-pane > li > div > div.m-module-bg > div.list > div.m-reviewCard-item",
			Attrs: map[string]CssSelector{
				"date": {
					Css: "div.comment > div > div.reviewDate",
				},
				"content": {
					Css: "div.comment > p",
				},
			},
		},
		Paginator: CssSelector{
			Css: "a.forward.active",
		},
		Limit: 5,
		EnableCookies: HttpCookies{
			RawCookies: "your cookies",
			Domain:     "hotels.ctrip.com",
		},
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

func TestRabidaImplPrePaginate1_Crawl(t *testing.T) {
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
		Limit: 10,
		EnableCookies: HttpCookies{
			RawCookies: "your cookies",
			Domain:     "you.ctrip.com",
		},
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
