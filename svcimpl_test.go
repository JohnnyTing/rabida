package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/JohnnyTing/rabida/lib"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	_ "github.com/unionj-cloud/go-doudou/svc/config"
	"log"
	"strings"
	"testing"
	"time"
)

func TestRabidaImpl_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://zjt.fujian.gov.cn/xxgk/zxwj/zxwj/",
		CssSelector: CssSelector{
			Scope: `.box>div>div:not([style="display: none;"])>div.gl_news`,
			Attrs: map[string]CssSelector{
				"title": {
					Css: ".gl_news_top_tit",
				},
				"link": {
					Css:  "a",
					Attr: "href",
				},
				"date": {
					Css: ".gl_news_top_rq",
				},
			},
		},
		Paginator: CssSelector{
			Css:  ".c-txt>a:nth-last-of-type(2)",
			Attr: "href",
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
		Limit: 10,
		EnableCookies: HttpCookies{
			RawCookies: "AHeadUserInfo=VipGrade=0&VipGradeName=%C6%D5%CD%A8%BB%E1%D4%B1&UserName=&NoReadMessageCount=0; DUID=u=754BF301C3EBA5CD980F210BDADB599E&v=0; IsNonUser=F; _RGUID=f17eceb0-45b2-4844-b1f1-cceba2627021; _RSG=4Xai51hp.5FX0dOYkEpn7B; _RDG=289e3e7b65b1772261218d2c5cbb85be06; _bfaStatusPVSend=1; login_uid=754BF301C3EBA5CD980F210BDADB599E; login_type=0; cticket=BF79374B1282A6DEC82AF3355D15DEB14555C2E1D69828F089021C63E4CE7E3C; _PRO_sso_lat_assertion_=541e53d70d876c073acd7bf54640243de43b8177accc89081d523b37c27701a28d640fc67f37b8a75e8c83b94334f3f1721cf500742762ff381482440180f7ba555e130fb52cc8071b5df5ee264a50fc68081cacbaf84fa3c50f893a5be3eb30c3e267da7ef9ccd724a63b1b19085cba7240233841aca7d3b95213c38bb6757e; _PRO_sso_lat_assertion_signature_=31906bbd8dbb73b781e8ee9aa4fde100892afc155ac7d687ca1d8b47791a70d3fa2c3b37cee43ecb35e71c9b944acc9d9a89a9c396fb92e33576f99e5262a82183a53f9e851c6bec326f904d4eaf2bb918ceaf3b58fb5b6841c7a312ec955209f1bd15af296d89383b8ca19aa5242c1fca26eaea6df273cb85f30fb045a67808; GUID=09031021411428148953; nfes_isSupportWebP=1; MKT_CKID=1638145516014.jy1ma.beis; MKT_CKID_LMT=1641371579056; _RF1=171.221.151.153; ibulanguage=CN; ibulocale=zh_cn; cookiePricesDisplayed=CNY; intl_ht1=h4=30_635264; MKT_Pagesource=PC; _jzqco=%7C%7C%7C%7C%7C1.1959893848.1641371579063.1641446083666.1641446120701.1641446083666.1641446120701.0.0.0.11.11; __zpspc=9.4.1641446083.1641446120.2%234%7C%7C%7C%7C%7C%23; appFloatCnt=7; librauuid=FLPs2lP5dPptuMnZ; _bfa=1.1640843640417.l4wk5.1.1641438245036.1641446080135.5.18; _bfs=1.4; _bfaStatus=send",
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
			RawCookies: "nfes_isSupportWebP=1; AHeadUserInfo=VipGrade=0&VipGradeName=%C6%D5%CD%A8%BB%E1%D4%B1&UserName=&NoReadMessageCount=0; DUID=u=754BF301C3EBA5CD980F210BDADB599E&v=0; IsNonUser=F; ASP.NET_SessionSvc=MTAuNjAuNDkuNzh8OTA5MHxqaW5xaWFvfGRlZmF1bHR8MTYzODQzMjEzODI3Mg; _RGUID=f17eceb0-45b2-4844-b1f1-cceba2627021; _RSG=4Xai51hp.5FX0dOYkEpn7B; _RDG=289e3e7b65b1772261218d2c5cbb85be06; _bfaStatusPVSend=1; login_uid=754BF301C3EBA5CD980F210BDADB599E; login_type=0; cticket=BF79374B1282A6DEC82AF3355D15DEB14555C2E1D69828F089021C63E4CE7E3C; _PRO_sso_lat_assertion_=541e53d70d876c073acd7bf54640243de43b8177accc89081d523b37c27701a28d640fc67f37b8a75e8c83b94334f3f1721cf500742762ff381482440180f7ba555e130fb52cc8071b5df5ee264a50fc68081cacbaf84fa3c50f893a5be3eb30c3e267da7ef9ccd724a63b1b19085cba7240233841aca7d3b95213c38bb6757e; _PRO_sso_lat_assertion_signature_=31906bbd8dbb73b781e8ee9aa4fde100892afc155ac7d687ca1d8b47791a70d3fa2c3b37cee43ecb35e71c9b944acc9d9a89a9c396fb92e33576f99e5262a82183a53f9e851c6bec326f904d4eaf2bb918ceaf3b58fb5b6841c7a312ec955209f1bd15af296d89383b8ca19aa5242c1fca26eaea6df273cb85f30fb045a67808; GUID=09031021411428148953; nfes_isSupportWebP=1; MKT_CKID=1638145516014.jy1ma.beis; MKT_CKID_LMT=1641371579056; _RF1=171.221.151.153; _pd=%7B%22r%22%3A1%2C%22d%22%3A801%2C%22_d%22%3A800%2C%22p%22%3A825%2C%22_p%22%3A24%2C%22o%22%3A846%2C%22_o%22%3A21%2C%22s%22%3A860%2C%22_s%22%3A14%7D; ibulanguage=CN; ibulocale=zh_cn; cookiePricesDisplayed=CNY; MKT_Pagesource=PC; intl_ht1=h4=31_1508003,30_635264; _bfa=1.1640843640417.l4wk5.1.1641438245036.1641446080135.5.20; _bfs=1.6; _jzqco=%7C%7C%7C%7C%7C1.1959893848.1641371579063.1641447992949.1641448268215.1641447992949.1641448268215.0.0.0.14.14; __zpspc=9.4.1641446083.1641448268.5%234%7C%7C%7C%7C%7C%23; appFloatCnt=10; _bfaStatus=fail; librauuid=4gO12nD96ndzVQU3",
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

func TestRabidaImplDosomethinBefore_Crawl(t *testing.T) {
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
		Limit: 10,
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

func TestRabidaXpathImpl_Crawl(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := NewRabida(conf)
	job := Job{
		Link: "http://dpc.wuxi.gov.cn/hdjl/dczj/index.shtml#",
		CssSelector: CssSelector{
			//Scope: `#zz > div`,
			XpathScope: `//*[@id="zz"]/div`,
			Attrs: map[string]CssSelector{
				"title": {
					//Css: "dl > dt > a",
					Xpath: "/dl/dt/a",
				},
				"link": {
					//Css:  "dl > dt > a",
					//Attr: "href",
					Xpath: `/dl/dt/a/@href`,
				},
				"date": {
					//Css: "ul > li:nth-child(1)",
					Xpath: `/ul/li[1]/text()[1]`,
				},
			},
		},
		Paginator: CssSelector{
			//Css: "a.next",
			//Xpath: "",
		},
		Limit: 10,
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
		panic(fmt.Sprintf("%+v", err))
	}
}

func TestRabidaImpl_WaitNewTarget(t *testing.T) {
	var (
		allocCancel   context.CancelFunc
		contextCancel context.CancelFunc
		ctx           context.Context
	)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"),
	}
	opts = append(opts, chromedp.Headless)
	ctx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	//ctx, contextCancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	ctx, contextCancel = chromedp.NewContext(ctx)
	defer contextCancel()

	downloadBegin := make(chan bool)
	var downloadUrl string

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*page.EventWindowOpen); ok {
			downloadUrl = ev.URL
			close(downloadBegin)
		}
	})

	var res []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("http://minzheng.hebei.gov.cn/policyDatabase"),
		chromedp.EvaluateAsDevTools("document.querySelector('.cell>a').click()", &res),
	); err != nil {
		log.Fatal(err)
	}

	// This will block until the chromedp listener closes the channel
	<-downloadBegin

	// We can predict the exact file location and name here because of how we configured
	// SetDownloadBehavior and WithDownloadPath
	log.Printf("Download url: %s", downloadUrl)
}

func TestRabidaImpl_CrawlWithListeners(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://minzheng.hebei.gov.cn/policyDatabase",
		CssSelector: CssSelector{
			Scope: `div.el-table__body-wrapper.is-scrolling-none > table > tbody > tr`,
			Attrs: map[string]CssSelector{
				"title": {
					Css: "td:first-child a",
				},
				"link": {
					Css:  "td:first-child a",
					Attr: "node",
				},
				"date": {
					Css: "td:last-child>div>div",
				},
			},
		},
		Paginator: CssSelector{
			Css: "button.btn-next>span",
		},
		Limit: 3,
	}

	linkCh := make(chan string, 1)
	err := rabi.CrawlWithListeners(context.Background(), job, func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			value, ok := item.(map[string]interface{})
			if !ok {
				panic(errors.New("cast failed"))
			}
			node := value["link"].(*cdp.Node)
			timeoutCtx, jsClickCancel := context.WithTimeout(ctx, 10*time.Second)
			err := chromedp.Run(timeoutCtx, lib.JsClickNode(node))
			if err != nil {
				jsClickCancel()
				panic(err)
			}
			jsClickCancel()
			link := <-linkCh
			log.Println(link)
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, []chromedp.Action{
		chromedp.EmulateViewport(1777, 903, chromedp.EmulateLandscape),
	}, nil,
		[]chromedp.ExecAllocatorOption{chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36")},
		func(v interface{}) {
			if ev, ok := v.(*page.EventWindowOpen); ok {
				linkCh <- ev.URL
			}
		},
	)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func TestRabidaImpl_CrawlWithListeners2(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://mfb.sh.gov.cn/zwgk/jcgk/zcfg/gfxwj/index.html",
		CssSelector: CssSelector{
			Scope: `#Datatable-1>tbody>tr`,
			Attrs: map[string]CssSelector{
				"title": {
					Css: "td:first-child",
				},
				"link": {
					Css:  "td:first-child a",
					Attr: "node",
				},
				"date": {
					Css: "td:last-child>div>div",
				},
			},
		},
		Paginator: CssSelector{
			Css: "button.btn-next>span",
		},
		Limit: 3,
	}

	linkCh := make(chan string, 1)
	err := rabi.CrawlWithListeners(context.Background(), job, func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			value, ok := item.(map[string]interface{})
			if !ok {
				panic(errors.New("cast failed"))
			}
			node := value["link"].(*cdp.Node)
			timeoutCtx, jsClickCancel := context.WithTimeout(ctx, 10*time.Second)
			err := chromedp.Run(timeoutCtx, lib.JsClickNode(node))
			if err != nil {
				jsClickCancel()
				panic(err)
			}
			jsClickCancel()
			link := <-linkCh
			log.Println(link)
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	}, nil, []chromedp.Action{
		chromedp.EmulateViewport(1777, 903, chromedp.EmulateLandscape),
	}, nil,
		[]chromedp.ExecAllocatorOption{chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36")},
		func(v interface{}) {
			if ev, ok := v.(*page.EventWindowOpen); ok {
				linkCh <- ev.URL
			}
		},
	)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func TestRabidaImpl_Download(t *testing.T) {
	var (
		allocCancel   context.CancelFunc
		contextCancel context.CancelFunc
		ctx           context.Context
	)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"),
	}
	opts = append(opts, chromedp.Headless)
	ctx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	//ctx, contextCancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	ctx, contextCancel = chromedp.NewContext(ctx)
	defer contextCancel()

	// set up a channel so we can block later while we monitor the download progress
	downloadComplete := make(chan bool)

	// this will be used to capture the file name later
	var fileName string

	// set up a listener to watch the download events and close the channel when complete
	// this could be expanded to handle multiple downloads through creating a guid map,
	// monitor download urls via EventDownloadWillBegin, etc
	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadWillBegin); ok {
			fileName = ev.SuggestedFilename
			return
		}
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			fmt.Printf("current download state: %s\n", ev.State.String())
			if ev.State == browser.DownloadProgressStateCompleted {
				close(downloadComplete)
			}
		}
	})

	if err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath("out").
			WithEventsEnabled(true),
		chromedp.Navigate("http://minzheng.hebei.gov.cn/jinge/Document/20211129/file20211129145312348.pdf"),
	); err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here since downloads
		// will cause this error to be emitted, although the download will still succeed.
		log.Fatal(err)
	}

	// This will block until the chromedp listener closes the channel
	<-downloadComplete

	log.Printf("Download Complete: %v/%v", "out", fileName)
}

func TestRabidaImpl_DownloadFile(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	job := Job{
		Link: "http://minzheng.hebei.gov.cn/jinge/Document/20211129/file20211129145312348.pdf",
	}
	err := rabi.DownloadFile(context.Background(), job, func(file string) {
		fmt.Println(file)
	}, conf)
	if err != nil {
		panic(err)
	}
}

func TestRabidaImpl_CrawlTraversal(t *testing.T) {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)
	rabi := NewRabida(conf)
	err := rabi.CrawlTraversal(context.Background(), conf)
	if err != nil {
		panic(err)
	}
}
