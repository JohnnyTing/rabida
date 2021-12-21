package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/fileutils"
	"github.com/unionj-cloud/go-doudou/pathutils"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"github.com/unionj-cloud/rabida/config"
	"github.com/unionj-cloud/rabida/lib"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

type RabidaImpl struct {
	conf *config.RabiConfig
}

func (r RabidaImpl) DownloadFile(ctx context.Context, job Job, callback func(file string), confPtr *config.RabiConfig, options ...chromedp.ExecAllocatorOption) error {
	var (
		err  error
		out  string
		conf config.RabiConfig
	)

	if confPtr != nil {
		conf = *confPtr
	} else {
		conf = *r.conf
	}

	if r.conf.Debug {
		out = r.conf.Out
		if out, err = pathutils.FixPath(out, ""); err != nil {
			return errors.Wrap(err, "")
		}
		if err = fileutils.CreateDirectory(out); err != nil {
			return errors.Wrap(err, "")
		}
	}

	if _ctx := chromedp.FromContext(ctx); _ctx == nil {
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
		}
		opts = append(opts, options...)
		if conf.Mode == "headless" {
			opts = append(opts, chromedp.Headless)
		}
		var (
			allocCancel   context.CancelFunc
			contextCancel context.CancelFunc
		)
		ctx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
		defer allocCancel()

		if r.conf.Debug {
			ctx, contextCancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
		} else {
			ctx, contextCancel = chromedp.NewContext(ctx)
		}
		defer contextCancel()
	}

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
			log.Printf("file %s current download state: %s\n", fileName, ev.State.String())
			if ev.State == browser.DownloadProgressStateCompleted {
				log.Printf("file %s download done\n", fileName)
				close(downloadComplete)
			}
		}
	})

	if err = chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(conf.Out).
			WithEventsEnabled(true),
		chromedp.Navigate(job.Link),
	); err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here since downloads
		// will cause this error to be emitted, although the download will still succeed.
		return errors.Wrap(err, "")
	}

	// This will block until the chromedp listener closes the channel
	<-downloadComplete

	log.Printf("Download Complete: %s\n", filepath.Join(conf.Out, fileName))
	callback(filepath.Join(conf.Out, fileName))
	return nil
}

func (r RabidaImpl) CrawlWithListeners(ctx context.Context, job Job, callback func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, confPtr *config.RabiConfig, options []chromedp.ExecAllocatorOption, listeners ...func(ev interface{})) error {
	var (
		err         error
		abort       bool
		ret         []interface{}
		nextPageUrl string
		pageNo      int
		timeoutCtx  context.Context
		out         string
		conf        config.RabiConfig
	)

	if confPtr != nil {
		conf = *confPtr
	} else {
		conf = *r.conf
	}

	if r.conf.Debug {
		out = r.conf.Out
		if out, err = pathutils.FixPath(out, ""); err != nil {
			return errors.Wrap(err, "")
		}
		if err = fileutils.CreateDirectory(out); err != nil {
			return errors.Wrap(err, "")
		}
	}

	if _ctx := chromedp.FromContext(ctx); _ctx == nil {
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
		}
		opts = append(opts, options...)
		if conf.Mode == "headless" {
			opts = append(opts, chromedp.Headless)
		}
		var (
			allocCancel   context.CancelFunc
			contextCancel context.CancelFunc
		)
		ctx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
		defer allocCancel()

		if r.conf.Debug {
			ctx, contextCancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
		} else {
			ctx, contextCancel = chromedp.NewContext(ctx)
		}
		defer contextCancel()
	}

	link := job.Link
	if stringutils.IsNotEmpty(job.StartPageUrl) {
		link = job.StartPageUrl
	}

	var tasks chromedp.Tasks

	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		_, err = page.AddScriptToEvaluateOnNewDocument(lib.Script).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}))

	tasks = append(tasks, before...)

	if err = chromedp.Run(ctx, tasks); err != nil {
		return errors.Wrap(err, "")
	}

	tasks = nil
	tasks = append(tasks, network.Enable(), lib.Navigate(link))
	tasks = append(tasks, after...)

	for _, fn := range listeners {
		chromedp.ListenTarget(ctx, fn)
	}

	if r.conf.Debug {
		chromedp.ListenTarget(ctx, func(event interface{}) {
			switch ev := event.(type) {
			case *runtime.EventConsoleAPICalled:
				logrus.Printf("* console.%s call:\n", ev.Type)
				for _, arg := range ev.Args {
					logrus.Printf("%s - %s\n", arg.Type, arg.Value)
				}
			case *runtime.EventExceptionThrown:
				// Since ts.URL uses a random port, replace it.
				s := ev.ExceptionDetails.Error()
				logrus.Printf("* %s\n", s)
			case *network.EventResponseReceived:
				if ev.Type == network.ResourceTypeXHR || ev.Type == network.ResourceTypeFetch {
					logrus.Println(gabs.Wrap(ev.Response).StringIndent("", "  "))
				}
			}
		})
	}

	var taskCancel context.CancelFunc
	timeoutCtx, taskCancel = context.WithTimeout(ctx, conf.Timeout)
	defer taskCancel()

	if err = chromedp.Run(timeoutCtx, tasks); err != nil {
		return errors.Wrap(err, "")
	}

	time.Sleep(conf.Timeout)

	startPageBtn := job.StartPageBtn.Css
	if stringutils.IsEmpty(startPageBtn) {
		startPageBtn = job.StartPageBtn.Xpath
	}
	if stringutils.IsNotEmpty(startPageBtn) {
		var father *cdp.Node
		if job.CssSelector.Iframe {
			if father, err = iframe(ctx, conf.Timeout); err != nil {
				return errors.Wrap(err, "")
			}
		}
		var cancel1 context.CancelFunc
		timeoutCtx, cancel1 = context.WithTimeout(ctx, conf.Timeout)
		defer cancel1()
		var buttons []*cdp.Node
		if father != nil {
			if err = chromedp.Run(timeoutCtx, chromedp.Nodes(startPageBtn, &buttons, chromedp.BySearch, chromedp.FromNode(father))); err != nil {
				return errors.Wrap(err, "")
			}
		} else {
			if err = chromedp.Run(timeoutCtx, chromedp.Nodes(startPageBtn, &buttons, chromedp.BySearch)); err != nil {
				return errors.Wrap(err, "")
			}
		}
		if len(buttons) > 0 {
			nextPageBtn := buttons[0]
			var cancel2 context.CancelFunc
			timeoutCtx, cancel2 = context.WithTimeout(ctx, conf.Timeout)
			defer cancel2()
			if err = chromedp.Run(timeoutCtx, lib.JsClickNode(nextPageBtn)); err != nil {
				return errors.Wrap(err, "")
			}
		}
		time.Sleep(conf.Timeout)
	}

	if err = prePaginate(ctx, job, conf); err != nil {
		return err
	}

	pageNo++
	if r.conf.Debug {
		if err = screenshot(ctx, out, pageNo); err != nil {
			return errors.Wrap(err, "")
		}
		if err = writeHtml(ctx, out, pageNo); err != nil {
			return errors.Wrap(err, "")
		}
	}

	ret, nextPageUrl, err = r.extract(ctx, job, conf)
	if err != nil {
		return errors.Wrap(err, "")
	}

	if abort = callback(ctx, ret, nextPageUrl, pageNo); abort {
		return nil
	}

	if stringutils.IsEmpty(job.Paginator.Css) && stringutils.IsEmpty(job.Paginator.Xpath) {
		return nil
	}

	r.sleep(conf)

	for {
		var father *cdp.Node
		if job.CssSelector.Iframe {
			if father, err = iframe(ctx, conf.Timeout); err != nil {
				return errors.Wrap(err, "")
			}
		}
		var nodeCancel context.CancelFunc
		timeoutCtx, nodeCancel = context.WithTimeout(ctx, conf.Timeout)
		var buttons []*cdp.Node
		pagination := job.Paginator.Css
		if stringutils.IsEmpty(pagination) {
			pagination = job.Paginator.Xpath
		}
		if father != nil {
			if err = chromedp.Run(timeoutCtx, chromedp.Nodes(pagination, &buttons, chromedp.BySearch, chromedp.FromNode(father))); err != nil {
				nodeCancel()
				goto ERR
			}
		} else {
			if err = chromedp.Run(timeoutCtx, chromedp.Nodes(pagination, &buttons, chromedp.BySearch)); err != nil {
				nodeCancel()
				goto ERR
			}
		}
		nodeCancel()
		if len(buttons) > 0 {
			nextPageBtn := buttons[0]
			var jsClickCancel context.CancelFunc
			timeoutCtx, jsClickCancel = context.WithTimeout(ctx, conf.Timeout)
			if err = chromedp.Run(timeoutCtx, lib.JsClickNode(nextPageBtn)); err != nil {
				jsClickCancel()
				goto ERR
			}
			jsClickCancel()
		}
		time.Sleep(conf.Timeout)
		pageNo++
		if r.conf.Debug {
			if err = screenshot(ctx, out, pageNo); err != nil {
				return errors.Wrap(err, "")
			}
			if err = writeHtml(ctx, out, pageNo); err != nil {
				return errors.Wrap(err, "")
			}
		}
		if ret, nextPageUrl, err = r.extract(ctx, job, conf); err != nil {
			goto ERR
		}
		if abort = callback(ctx, ret, nextPageUrl, pageNo); abort {
			goto END
		}
		r.sleep(conf)
		continue

	END:
		return nil
	ERR:
		if errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return errors.Wrap(err, "")
	}
}

func screenshot(ctx context.Context, out string, pageNo int) (err error) {
	var buf []byte
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = chromedp.Run(timeoutCtx, chromedp.FullScreenshot(&buf, 100)); err != nil {
		return errors.Wrap(err, "")
	}
	if err = ioutil.WriteFile(filepath.Join(out, fmt.Sprintf("screenshot_%d.png", pageNo)), buf, 0644); err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func writeHtml(ctx context.Context, out string, pageNo int) (err error) {
	var html string
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = chromedp.Run(timeoutCtx, chromedp.OuterHTML("html", &html, chromedp.ByQuery)); err != nil {
		return errors.Wrap(err, "")
	}
	if err = ioutil.WriteFile(filepath.Join(out, fmt.Sprintf("index_%d.html", pageNo)), []byte(html), 0644); err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func prePaginate(ctx context.Context, job Job, conf config.RabiConfig) error {
	if stringutils.IsNotEmpty(string(job.PrePaginate.Type)) {
		timeoutCtx, nodeCancel := context.WithTimeout(ctx, conf.Timeout)
		defer nodeCancel()
		prePaginateSelector := job.PrePaginate.Selector.Css
		if stringutils.IsEmpty(prePaginateSelector) {
			prePaginateSelector = job.PrePaginate.Selector.Xpath
		}
		if stringutils.IsNotEmpty(prePaginateSelector) {
			switch job.PrePaginate.Type {
			case ClickEvent:
				if err := chromedp.Run(timeoutCtx, chromedp.Click(prePaginateSelector, chromedp.BySearch)); err != nil {
					return errors.Wrap(err, "PrePaginate Error")
				}
			}
		}
	}
	return nil
}

func (r RabidaImpl) CrawlWithConfig(ctx context.Context, job Job, callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, conf config.RabiConfig, options ...chromedp.ExecAllocatorOption) error {
	return r.CrawlWithListeners(ctx, job, func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		return callback(ret, nextPageUrl, currentPageNo)
	}, before, after, &conf, options)
}

func iframe(ctx context.Context, timeout time.Duration) (iframe *cdp.Node, err error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var iframes []*cdp.Node
	if err = chromedp.Run(timeoutCtx, chromedp.Nodes("iframe", &iframes, chromedp.ByQuery)); err != nil {
		return nil, errors.Wrap(err, "")
	}
	if len(iframes) > 0 {
		iframe = iframes[0]
	}
	return
}

func (r RabidaImpl) sleep(conf config.RabiConfig) {
	var s time.Duration
	if len(conf.Delay) > 1 {
		s = lib.RandDuration(conf.Delay[0], conf.Delay[1])
	} else {
		s = conf.Delay[0]
	}
	logrus.Infof("sleep %s to crawl next page\n", s.String())
	time.Sleep(s)
}

func (r RabidaImpl) Crawl(ctx context.Context, job Job, callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool,
	before []chromedp.Action, after []chromedp.Action) error {
	return r.CrawlWithConfig(ctx, job, callback, before, after, *r.conf)
}

type errNotFound struct{}

func (e errNotFound) Error() string {
	return "not found"
}

var ErrNotFound error = errNotFound{}

func (r RabidaImpl) populate(ctx context.Context, father *cdp.Node, cssSelector CssSelector, conf config.RabiConfig) []interface{} {
	scope := cssSelector.Scope
	if stringutils.IsEmpty(scope) && father == nil {
		scope = "html"
	}
	var nodes []*cdp.Node
	if stringutils.IsNotEmpty(scope) {
		timeoutCtx, nodeCancel := context.WithTimeout(ctx, conf.Timeout)
		defer nodeCancel()
		if father != nil {
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll, chromedp.FromNode(father))); err != nil {
				logrus.Error(fmt.Sprintf("%+v", errors.Wrap(ErrNotFound, scope)))
			}
		} else {
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll)); err != nil {
				logrus.Error(fmt.Sprintf("%+v", errors.Wrap(ErrNotFound, scope)))
			}
		}
	} else {
		nodes = append(nodes, father)
	}
	var ret []interface{}
	for _, node := range nodes {
		if cssSelector.Attrs == nil {
			timeoutCtx, attrCancel := context.WithTimeout(ctx, conf.Timeout)
			var value interface{}
			if stringutils.IsEmpty(cssSelector.Attr) {
				if stringutils.IsEmpty(cssSelector.Css) {
					_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute([]cdp.NodeID{node.NodeID}, "innerText", &value, chromedp.ByNodeID))
				} else {
					var stringValue string
					var _nodes []*cdp.Node
					_ = chromedp.Run(timeoutCtx, chromedp.Nodes(cssSelector.Css, &_nodes, chromedp.ByQueryAll, chromedp.FromNode(node)))
					for _, _node := range _nodes {
						var temp string
						_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute([]cdp.NodeID{_node.NodeID}, "innerText", &temp, chromedp.ByNodeID))
						stringValue += temp
					}
					value = stringValue
				}
			} else {
				var stringValue string
				if stringutils.IsEmpty(cssSelector.Css) {
					if cssSelector.Attr == "outerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.OuterHTML([]cdp.NodeID{node.NodeID}, &stringValue, chromedp.ByNodeID))
						value = stringValue
					} else if cssSelector.Attr == "innerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.InnerHTML([]cdp.NodeID{node.NodeID}, &stringValue, chromedp.ByNodeID))
						value = stringValue
					} else if cssSelector.Attr == "node" {
						value = node
					} else {
						_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute([]cdp.NodeID{node.NodeID}, cssSelector.Attr, &value, chromedp.ByNodeID))
					}
				} else {
					if cssSelector.Attr == "outerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.OuterHTML(cssSelector.Css, &stringValue, chromedp.ByQuery, chromedp.FromNode(node)))
						value = stringValue
					} else if cssSelector.Attr == "innerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.InnerHTML(cssSelector.Css, &stringValue, chromedp.ByQuery, chromedp.FromNode(node)))
						value = stringValue
					} else if cssSelector.Attr == "innerText" {
						var _nodes []*cdp.Node
						_ = chromedp.Run(timeoutCtx, chromedp.Nodes(cssSelector.Css, &_nodes, chromedp.ByQueryAll, chromedp.FromNode(node)))
						for _, _node := range _nodes {
							var temp string
							_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute([]cdp.NodeID{_node.NodeID}, "innerText", &temp, chromedp.ByNodeID))
							stringValue += temp
						}
						value = stringValue
					} else if cssSelector.Attr == "node" {
						var _nodes []*cdp.Node
						_ = chromedp.Run(timeoutCtx, chromedp.Nodes(cssSelector.Css, &_nodes, chromedp.ByQuery, chromedp.FromNode(node)))
						if len(_nodes) > 0 {
							value = _nodes[0]
						}
					} else {
						_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(cssSelector.Css, cssSelector.Attr, &value, chromedp.ByQuery, chromedp.FromNode(node)))
					}
				}
			}
			if value != nil && !reflect.ValueOf(value).IsZero() {
				ret = append(ret, value)
			}
			attrCancel()
		} else {
			data := make(map[string]interface{})
			for attr, sel := range cssSelector.Attrs {
				result := r.populate(ctx, node, sel, conf)
				if len(result) > 0 {
					if stringutils.IsEmpty(sel.Scope) {
						data[attr] = result[0]
					} else {
						data[attr] = result
					}
				}
			}
			if len(data) == 0 {
				continue
			}
			ret = append(ret, data)
		}
	}
	return ret
}

func (r RabidaImpl) populateX(ctx context.Context, cssSelector CssSelector, conf config.RabiConfig, html *html.Node) []interface{} {
	var ret []interface{}
	if stringutils.IsNotEmpty(cssSelector.XpathScope) {
		nodes := htmlquery.Find(html, cssSelector.XpathScope)
		for _, node := range nodes {
			ret = r.recursivePopulateX(ctx, cssSelector, conf, node, ret)
		}
	} else {
		ret = r.recursivePopulateX(ctx, cssSelector, conf, html, ret)
	}
	return ret
}

func (r RabidaImpl) recursivePopulateX(ctx context.Context, cssSelector CssSelector, conf config.RabiConfig, node *html.Node, ret []interface{}) []interface{} {
	if cssSelector.Attrs == nil {
		value := retrieveByXpath(ctx, cssSelector, node)
		if stringutils.IsNotEmpty(value) {
			ret = append(ret, value)
		}
	} else {
		data := make(map[string]interface{})
		for attr, sel := range cssSelector.Attrs {
			result := r.populateX(ctx, sel, conf, node)
			if len(result) > 0 {
				if stringutils.IsEmpty(sel.XpathScope) {
					data[attr] = result[0]
				} else {
					data[attr] = result
				}
			}
		}
		if len(data) > 0 {
			ret = append(ret, data)
		}
	}
	return ret
}

func retrieveByXpath(ctx context.Context, cssSelector CssSelector, html *html.Node) (value string) {
	if stringutils.IsNotEmpty(cssSelector.XpathScope) {
		nodes := htmlquery.Find(html, cssSelector.XpathScope)
		for _, _node := range nodes {
			value += lib.FindOne(_node, cssSelector.Xpath)
		}
	} else {
		value = lib.FindOne(html, cssSelector.Xpath)
	}
	return
}

func (r RabidaImpl) extract(ctx context.Context, job Job, conf config.RabiConfig) (ret []interface{}, nextPageUrl string, err error) {
	defer func() {
		if val := recover(); val != nil {
			var ok bool
			err, ok = val.(error)
			if !ok {
				err = errors.New(fmt.Sprint(val))
			} else {
				err = errors.Wrap(err, "recover from panic")
			}
		}
	}()

	var father *cdp.Node
	if job.CssSelector.Iframe {
		if father, err = iframe(ctx, conf.Timeout); err != nil {
			panic(errors.Wrap(err, ""))
		}
	}

	if stringutils.IsNotEmpty(job.CssSelector.XpathScope) || stringutils.IsNotEmpty(job.CssSelector.Xpath) {
		doc := r.Html(ctx, father, conf)
		ret = r.populateX(ctx, job.CssSelector, conf, doc)
		if stringutils.IsNotEmpty(job.Paginator.Xpath) {
			nextPageUrl = lib.FindOne(doc, job.Paginator.Xpath)
		}
	} else {
		ret = r.populate(ctx, father, job.CssSelector, conf)
		if stringutils.IsNotEmpty(job.Paginator.Css) && stringutils.IsNotEmpty(job.Paginator.Attr) {
			timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
			defer cancel()
			_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(job.Paginator.Css, job.Paginator.Attr, &nextPageUrl, chromedp.ByQuery))
		}
	}
	return
}

func (r RabidaImpl) Html(ctx context.Context, father *cdp.Node, conf config.RabiConfig) *html.Node {
	var root string
	timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	if father == nil {
		if err := chromedp.Run(timeoutCtx, chromedp.OuterHTML("html", &root)); err != nil {
			logrus.Error(fmt.Sprintf("%+v", errors.Wrap(err, "")))
		}
	} else {
		if err := chromedp.Run(timeoutCtx, chromedp.OuterHTML("html", &root, chromedp.FromNode(father))); err != nil {
			logrus.Error(fmt.Sprintf("%+v", errors.Wrap(err, "")))
		}
	}
	doc, err := htmlquery.Parse(strings.NewReader(root))
	if err != nil {
		panic(errors.Wrap(err, "parse html error"))
	}
	return doc
}

func NewRabida(conf *config.RabiConfig) Rabida {
	return &RabidaImpl{
		conf,
	}
}
