package main

import (
	"context"
	"fmt"
	_ "github.com/unionj-cloud/go-doudou/svc/config"
	service "github.com/unionj-cloud/rabida"
	"github.com/unionj-cloud/rabida/config"
)

func main() {
	conf := config.LoadFromEnv()
	fmt.Printf("%+v\n", conf)

	rabi := service.NewRabida(conf)
	job := service.Job{
		Link:  "http://zjt.fujian.gov.cn/xxgk/zxwj/zxwj/",
		Scope: `.box>div>div:not([style="display: none;"])>div.gl_news`,
		Attrs: map[string]service.CssSelector{
			"title": service.CssSelector{
				Css: ".gl_news_top_tit",
			},
			"link": service.CssSelector{
				Css:  "a",
				Attr: "href",
			},
			"date": service.CssSelector{
				Css: ".gl_news_top_rq",
			},
		},
		Paginator: ".c-txt>a:nth-last-of-type(2)",
		Limit:     20,
	}
	err := rabi.Crawl(context.Background(), job, func(ret []map[string]string, nextPageUrl string, currentPageNo int) bool {
		for _, item := range ret {
			fmt.Println(item["title"] + " " + item["link"])
		}
		if currentPageNo >= job.Limit {
			return true
		}
		return false
	})
	if err != nil {
		panic(err)
	}
}
