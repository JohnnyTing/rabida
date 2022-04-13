package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestRabidaXpathImpl_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := NewRabida(conf)
	job := Job{
		Link: "https://you.ctrip.com/sight/shenzhen26/2778.html",
		CssSelector: CssSelector{
			XpathScope: `//*[@id="commentModule"]/div[@class='commentList']/div`,
			Attrs: map[string]CssSelector{
				"content": {
					Xpath: "//div[@class='commentDetail']",
				},
				"date": {
					Xpath: `//div[@class='commentTime']`,
				},
			},
		},
		Paginator: CssSelector{
			Xpath: "//*[@id='commentModule']//li[@class=' ant-pagination-next' and not(@aria-disabled='true')]",
		},
		Limit: 3,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
		}
		logrus.Printf("currentPageNo: %d\n", currentPageNo)
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
