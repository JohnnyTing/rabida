### Rabida

Rabida 是一个基于 [chromedp](https://github.com/chromedp/chromedp) 简单易用的爬虫框架。

### 目前支持的特性

- `分页`:  使用css 选择器获取下页数据。
- `预分页`: 在获取分页数据前做一些操作，比如点击日期按钮获取最新数据。
- `浏览器cookie`: 启用浏览器cookie，针对于某些网站需要登录状态。
- `延迟跟超时`:  能够自定义延迟跟超时时间。
- `反爬虫检测`: 每个任务默认加载了反爬虫检测脚本，脚本来源于[puppeteer-extra-stealth](https://github.com/berstend/puppeteer-extra/tree/master/packages/extract-stealth-evasions#readme)。
- `严格模式`: useragent、浏览器、浏览器的平台必须匹配，如果设置成true，将设置为chrome-mac相关的useragent、chrome浏览器、浏览器平台为Mac。针对于某些网站的反爬机制。

### 安装

```go
go get -u github.com/JohnnyTing/rabida
```

### 配置

添加.env 文件到你的项目

```shell
RABI_DELAY=1s,2s
RABI_CONCURRENCY=1
RABI_THROTTLE_NUM=2
RABI_THROTTLE_DURATION=1s
RABI_TIMEOUT=3s
RABI_MODE=headless
RABI_DEBUG=false
RABI_OUT=out
RABI_STRICT=false
RABI_PROXY=
```

### 用法

这里看更多的例子 [examples](https://github.com/JohnnyTing/rabida/blob/master/examples) 

```go
func TestRabidaImplCrawl(t *testing.T) {
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
		panic(fmt.Sprintf("%+v", err))
	}
}
```