package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/JohnnyTing/rabida/lib"
	"github.com/JohnnyTing/rabida/useragent"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"log"
	"testing"
	"time"
)

func TestRabidaImplCrawl(t *testing.T) {
	t.Run("tieba.baidu.com", func(t *testing.T) {
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
			t.Errorf("%+v", err)
		}

	})
}

func TestRabidaImplCrawl_Twitter(t *testing.T) {
	t.Run("tieba.baidu.com", func(t *testing.T) {
		conf := config.LoadFromEnv()
		fmt.Printf("%+v\n", conf)
		rabi := NewRabida(conf)
		job := Job{
			Link: "https://twitter.com/NASA",
			CssSelector: CssSelector{
				Scope: `div[data-testid='cellInnerDiv'] article[data-testid='tweet']`,
				Attrs: map[string]CssSelector{
					"content": {
						Css: `div[data-testid="tweetText"]`,
					},
					"date": {
						Css:  "a > time",
						Attr: "datetime",
					},
				},
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
			t.Errorf("%+v", err)
		}

	})
}

func TestRabidaImplCrawl_TwitterScroll(t *testing.T) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(useragent.RandomMacChromeUA()))
	opts = append(opts, chromedp.Flag("headless", false))
	ctx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	var tasks chromedp.Tasks
	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		_, err = page.AddScriptToEvaluateOnNewDocument(lib.Script).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}))
	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		_, err = page.AddScriptToEvaluateOnNewDocument(lib.AntiDetectionJS).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}))

	tasks = append(tasks,
		chromedp.Navigate("https://twitter.com/NASA"),
	)

	if err := chromedp.Run(ctx, tasks); err != nil {
		t.Error(fmt.Sprintf("%+v", err))
	}
	for i := 0; i < 10; i++ {
		Scroll(ctx)
	}

}

func Scroll(ctx context.Context) {
	log.Println("start scroll")
	time.Sleep(time.Second)
	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes("div[data-testid='cellInnerDiv'] article[data-testid='tweet']", &nodes)); err != nil {
		log.Panicln(fmt.Sprintf("%+v", err))
	}
	log.Println(len(nodes))
	for _, node := range nodes {
		var (
			text string
			date string
			ok   bool
		)
		if err := chromedp.Run(ctx, chromedp.Text(`div[data-testid="tweetText"]`, &text, chromedp.FromNode(node))); err != nil {
			log.Panicln(fmt.Sprintf("%+v", err))
		}

		if err := chromedp.Run(ctx, chromedp.AttributeValue(`a > time`, "datetime", &date, &ok, chromedp.FromNode(node))); err != nil {
			log.Panicln(fmt.Sprintf("%+v", err))
		}
		data := make(map[string]string, 2)
		data["content"] = text
		data["date"] = date
		fmt.Println(gabs.Wrap(data).StringIndent("", "  "))
	}
	//document.body.scrollHeight
	//window.scrollTo(0,%v);

	js := fmt.Sprintf("window.scrollBy({top: 400, left: 100, behavior: 'smooth'});")
	//js := fmt.Sprintf("window.scrollBy(0,%v);", curor)
	// scroll
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_, exp, err := runtime.Evaluate(js).Do(ctx)
		if err != nil {
			return err
		}
		if exp != nil {
			return exp
		}
		return nil
	})); err != nil {
		log.Panicln(fmt.Sprintf("%+v", err))
	}
}
