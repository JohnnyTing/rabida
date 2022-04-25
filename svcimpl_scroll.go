package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/JohnnyTing/rabida/lib"
	"github.com/JohnnyTing/rabida/useragent"
	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/toolkit/fileutils"
	"github.com/unionj-cloud/go-doudou/toolkit/pathutils"
	"github.com/unionj-cloud/go-doudou/toolkit/stringutils"
	"golang.org/x/net/html"
	"reflect"
	"time"
)

func (r RabidaImpl) CrawlScroll(ctx context.Context, job Job, callback func(ret []interface{}, cursor int, currentPageNo int) bool,
	before []chromedp.Action, after []chromedp.Action) error {
	return r.CrawlScrollWithConfig(ctx, job, callback, before, after, *r.conf)
}

func (r RabidaImpl) CrawlScrollWithConfig(ctx context.Context, job Job, callback func(ret []interface{}, cursor int, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, conf config.RabiConfig, options ...chromedp.ExecAllocatorOption) error {
	return r.CrawlScrollWithListeners(ctx, job, func(ctx context.Context, ret []interface{}, cursor int, currentPageNo int) bool {
		return callback(ret, cursor, currentPageNo)
	}, before, after, &conf, options)
}

func (r RabidaImpl) CrawlScrollWithListeners(ctx context.Context, job Job, callback func(ctx context.Context, ret []interface{}, cursor int, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, confPtr *config.RabiConfig, options []chromedp.ExecAllocatorOption, listeners ...func(ev interface{})) error {
	var (
		err         error
		abort       bool
		ret         []interface{}
		cursor      int
		originScope string
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
			chromedp.Flag("ignore-certificate-errors", "1"),
		}
		var userAgent string
		if conf.Strict {
			userAgent = useragent.RandomMacChromeUA()
		} else {
			userAgent = useragent.RandomPcUA()
		}
		logrus.Infoln(userAgent)
		opts = append(opts, chromedp.UserAgent(userAgent))
		opts = append(opts, options...)
		if conf.Mode == "headless" {
			opts = append(opts, chromedp.Headless)
		}
		if stringutils.IsNotEmpty(conf.Proxy) {
			opts = append(opts, chromedp.ProxyServer(conf.Proxy))
		}
		var (
			allocCancel   context.CancelFunc
			contextCancel context.CancelFunc
		)
		ctx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
		defer allocCancel()

		if r.conf.Debug {
			ctx, contextCancel = chromedp.NewContext(ctx)
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

	if conf.Strict {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			_, err = page.AddScriptToEvaluateOnNewDocument(lib.AntiDetectionJS).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}))
	}

	tasks = append(tasks, before...)

	if err = chromedp.Run(ctx, tasks); err != nil {
		return errors.Wrap(err, "")
	}

	tasks = nil
	tasks = append(tasks, network.Enable())
	if stringutils.IsNotEmpty(job.EnableCookies.RawCookies) {
		tasks = append(tasks, lib.CookieAction(link, job.EnableCookies.RawCookies, job.EnableCookies.Expires))
	}
	tasks = append(tasks, lib.Navigate(link))
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

	var cancel context.CancelFunc
	timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
	defer cancel()

	if err = chromedp.Run(timeoutCtx, tasks); err != nil {
		return errors.Wrap(err, "")
	}

	DelaySleep(conf, "start run")

	if r.conf.Debug {
		timeoutCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err = chromedp.Run(timeoutCtx, chromedp.EvaluateAsDevTools(lib.DebugPrint, nil)); err != nil {
			return errors.Wrap(err, "")
		}
		chromedp.Sleep(10 * time.Second)
		if err = screenshot(ctx, out, -1); err != nil {
			return errors.Wrap(err, "")
		}
		if err = writeHtml(ctx, out, -1); err != nil {
			return errors.Wrap(err, "")
		}
		logrus.Infoln("index html screenshot wait 10s")
		time.Sleep(10 * time.Second)
	}

	startPageBtn := job.StartPageBtn.Css
	if stringutils.IsEmpty(startPageBtn) {
		startPageBtn = job.StartPageBtn.Xpath
	}
	if stringutils.IsNotEmpty(startPageBtn) {
		var father *cdp.Node
		if job.CssSelector.Iframe {
			if father, err = iframe(ctx, conf.Timeout, job); err != nil {
				return errors.Wrap(err, "")
			}
		}
		timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
		defer cancel()
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
			timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
			defer cancel()
			if err = chromedp.Run(timeoutCtx, lib.JsClickNode(nextPageBtn)); err != nil {
				return errors.Wrap(err, "")
			}
		}
		DelaySleep(conf, "startPageBtn")
	}

	if err = prePaginate(ctx, job, conf); err != nil {
		return err
	}

	pageNo++
	logrus.Infof("\n current pageNo: %v \n", pageNo)
	if r.conf.Debug {
		if err = screenshot(ctx, out, pageNo); err != nil {
			return errors.Wrap(err, "")
		}
		if err = writeHtml(ctx, out, pageNo); err != nil {
			return errors.Wrap(err, "")
		}
	}

	originScope = job.CssSelector.Scope
	if stringutils.IsEmpty(originScope) {
		originScope = job.CssSelector.XpathScope
	}
	logrus.Printf("scope: %s \n", originScope)

	ret, cursor, err = r.extractScroll(ctx, job, pageNo, conf, cursor)
	if err != nil {
		return errors.Wrap(err, "")
	}

	if abort = callback(ctx, ret, cursor, pageNo); abort {
		return nil
	}

	p := r.paginator(job, pageNo)
	if stringutils.IsEmpty(p.Css) && stringutils.IsEmpty(p.Xpath) {
		return nil
	}
	r.sleep(conf)

	old := chromedp.FromContext(ctx).Target.TargetID
	targetID := chromedp.FromContext(ctx).Target.TargetID
	var lazyCancel context.CancelFunc
	for {
		var keepOn bool
		var goon bool
		keepOn, err = func() (bool, error) {
			if lazyCancel != nil {
				defer lazyCancel()
			}
			newCtx := ctx
			if targetID != old {
				newCtx, cancel = chromedp.NewContext(ctx, chromedp.WithTargetID(targetID))
				defer cancel()
			}
			ch := chromedp.WaitNewTarget(newCtx, func(info *target.Info) bool {
				return info.URL != "" && info.URL != "about:blank"
			})
			var father *cdp.Node
			var buttons []*cdp.Node
			p = r.paginator(job, pageNo)
			pagination := p.Css
			if job.CssSelector.Iframe {
				if father, err = iframe(newCtx, conf.Timeout, job); err != nil {
					goto ERR
				}
			}
			if stringutils.IsEmpty(pagination) {
				pagination = p.Xpath
			}
			if goon, err = paginateCondition(newCtx, conf, job, father); err != nil {
				goto ERR
			}
			if !goon {
				goto END
			}

			timeoutCtx, cancel = context.WithTimeout(newCtx, conf.Timeout)
			defer cancel()
			if father != nil {
				if err = chromedp.Run(timeoutCtx, chromedp.Nodes(pagination, &buttons, chromedp.BySearch, chromedp.FromNode(father))); err != nil {
					goto ERR
				}
			} else {
				if err = chromedp.Run(timeoutCtx, chromedp.Nodes(pagination, &buttons, chromedp.BySearch)); err != nil {
					goto ERR
				}
			}
			if len(buttons) > 0 {
				nextPageBtn := buttons[0]
				timeoutCtx, cancel = context.WithTimeout(newCtx, conf.Timeout)
				defer cancel()
				if err = chromedp.Run(timeoutCtx, lib.JsClickNode(nextPageBtn)); err != nil {
					goto ERR
				}
				select {
				case targetID = <-ch:
					newCtx, lazyCancel = chromedp.NewContext(ctx, chromedp.WithTargetID(targetID))
				case <-time.After(1 * time.Second):
				}
			}
			DelaySleep(conf, "click next page")
			pageNo++
			logrus.Infof("\n current pageNo: %v \n", pageNo)
			if r.conf.Debug {
				if err = screenshot(newCtx, out, pageNo); err != nil {
					goto ERR
				}
				if err = writeHtml(newCtx, out, pageNo); err != nil {
					goto ERR
				}
			}

			if stringutils.IsNotEmpty(job.CssSelector.Scope) {
				job.CssSelector.Scope = fmt.Sprintf("%s:nth-child(n+%v)", originScope, cursor+1)
				logrus.Printf("scope: %s \n", job.CssSelector.Scope)
			} else if stringutils.IsNotEmpty(job.CssSelector.XpathScope) {
				job.CssSelector.XpathScope = lib.CursorScopeByPosition(originScope, cursor+1)
				logrus.Printf("xpath scope: %s \n", job.CssSelector.XpathScope)
			} else {
				logrus.Error("xpath scope css or xpath is empty")
				goto ERR
			}

			if ret, cursor, err = r.extractScroll(newCtx, job, pageNo, conf, cursor); err != nil {
				goto ERR
			}

			if abort = callback(newCtx, ret, cursor, pageNo); abort {
				goto END
			}

			r.sleep(conf)
			return true, nil

		END:
			return false, nil
		ERR:
			logrus.Error(errors.Wrap(err, ""))
			return false, err
		}()
		if !keepOn {
			return err
		}
	}
}

func (r RabidaImpl) extractScroll(ctx context.Context, job Job, pageNo int, conf config.RabiConfig, cursor int) (ret []interface{}, nextCursor int, err error) {
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
		if father, err = iframe(ctx, conf.Timeout, job); err != nil {
			panic(errors.Wrap(err, ""))
		}
	}

	DelaySleep(conf, "populate")

	if stringutils.IsNotEmpty(job.CssSelector.XpathScope) || stringutils.IsNotEmpty(job.CssSelector.Xpath) {
		doc := r.Html(ctx, father, conf)
		ret, nextCursor = r.populateXScroll(ctx, job.CssSelector, conf, doc)
	} else {
		ret, nextCursor = r.populateScroll(ctx, father, job.CssSelector, conf)
	}
	nextCursor += cursor
	return
}

func (r RabidaImpl) populateScroll(ctx context.Context, father *cdp.Node, cssSelector CssSelector, conf config.RabiConfig) (ret []interface{}, nextCursor int) {
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
				logrus.Error(fmt.Sprintf("scope err: %+v", errors.Wrap(ErrNotFound, scope)))
			}
		} else {
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll)); err != nil {
				logrus.Error(fmt.Sprintf("scope err: %+v", errors.Wrap(ErrNotFound, scope)))
			}
		}
		nextCursor = len(nodes)
	} else {
		nodes = append(nodes, father)
	}
	for _, node := range nodes {
		if cssSelector.Attrs == nil {
			err := doSomethingBefore(ctx, conf, cssSelector.Before, node)
			if err != nil {
				panic(err)
			}
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
				result, _ := r.populateScroll(ctx, node, sel, conf)
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
	return
}

func (r RabidaImpl) populateXScroll(ctx context.Context, cssSelector CssSelector, conf config.RabiConfig, html *html.Node) (ret []interface{}, cursor int) {
	if stringutils.IsNotEmpty(cssSelector.XpathScope) {
		nodes := htmlquery.Find(html, cssSelector.XpathScope)
		cursor = len(nodes)
		for _, node := range nodes {
			ret = r.recursivePopulateX(ctx, cssSelector, conf, node, ret)
		}
	} else {
		ret = r.recursivePopulateX(ctx, cssSelector, conf, html, ret)
	}
	return
}
