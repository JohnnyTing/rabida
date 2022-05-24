package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"testing"
)

func TestRabidaImplCrawlScrollSmooth(t *testing.T) {
	t.Run("CrawlScrollSmooth", func(t *testing.T) {
		conf := config.LoadFromEnv()
		fmt.Printf("%+v\n", conf)
		rabi := NewRabida(conf)
		job := Job{
			Link: "https://twitter.com/NASA",
			CssSelector: CssSelector{
				Scope: `div[data-testid='cellInnerDiv'] article[data-testid='tweet']`,
				Attrs: map[string]CssSelector{
					"title": {
						Css: `div[data-testid="tweetText"]`,
					},
					"date": {
						Css:  `a > time`,
						Attr: `datetime`,
					},
					"link": {
						Css:  `a[role="link"][href*=status]`,
						Attr: `href`,
					},
					"reply": {
						Css:  `div[data-testid="reply"]`,
						Attr: `aria-label`,
					},
					"retweet": {
						Css:  `div[data-testid="retweet"]`,
						Attr: `aria-label`,
					},
					"like": {
						Css:  `div[data-testid="like"]`,
						Attr: `aria-label`,
					},
				},
			},
			Limit: 5,
		}
		err := rabi.CrawlScrollSmooth(context.Background(), job, func(ret []interface{}, currentPageNo int) bool {
			for _, item := range ret {
				fmt.Println(gabs.Wrap(item).StringIndent("", "  "))
			}
			if currentPageNo >= job.Limit {
				return true
			}
			return false
		}, nil, nil)
		if err != nil {
			t.Errorf("%+v", err)
		}

	})
}
