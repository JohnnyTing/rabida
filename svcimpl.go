package service

import (
	"context"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/JohnnyTing/rabida/config"
	"github.com/JohnnyTing/rabida/lib"
	"github.com/JohnnyTing/rabida/useragent"
	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/browser"
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

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-downloadComplete:
	}

	log.Printf("Download Complete: %s\n", filepath.Join(conf.Out, fileName))
	callback(filepath.Join(conf.Out, fileName))
	return nil
}

func (r RabidaImpl) paginator(job Job, pageNo int) CssSelector {
	p := job.Paginator
	if job.PaginatorFunc != nil {
		p = job.PaginatorFunc(pageNo)
	}
	selector := p.Css
	if stringutils.IsEmpty(selector) {
		selector = p.Xpath
	}
	logrus.Infof("next page button selector is %s\n", selector)
	return p
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
			//ctx, contextCancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
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

	ret, nextPageUrl, err = r.extract(ctx, job, pageNo, conf)
	if err != nil {
		return errors.Wrap(err, "")
	}

	if abort = callback(ctx, ret, nextPageUrl, pageNo); abort {
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
			if ret, nextPageUrl, err = r.extract(newCtx, job, pageNo, conf); err != nil {
				goto ERR
			}
			if abort = callback(newCtx, ret, nextPageUrl, pageNo); abort {
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
	return doSomethingBefore(ctx, conf, job.PrePaginate, nil)
}

func CssOrXpath(cssSelector CssSelector) string {
	if stringutils.IsNotEmpty(cssSelector.Css) {
		return cssSelector.Css
	}
	return cssSelector.Xpath
}

func doSomethingBefore(ctx context.Context, conf config.RabiConfig, events []EventSelector, node *cdp.Node) error {
	if len(events) == 0 {
		return nil
	}
	var queryActions []chromedp.QueryOption
	// xpath expr is invalid: chromedp.BySearch
	queryActions = append(queryActions, chromedp.ByQuery)
	if node != nil {
		queryActions = append(queryActions, chromedp.FromNode(node))
	}
	for _, event := range events {
		if stringutils.IsNotEmpty(string(event.Type)) {
			DelaySleep(conf, "preprocess")
			timeoutCtx, nodeCancel := context.WithTimeout(ctx, conf.Timeout)
			defer nodeCancel()
			preSelectors := CssOrXpath(event.Selector)
			if stringutils.IsNotEmpty(preSelectors) {
				flag, err := ExecEventCondition(ctx, conf, &event.Condition, queryActions)
				if err != nil {
					return errors.Wrap(err, "event exc condition Error")
				}
				switch event.Type {
				case ClickEvent:
					if flag {
						if err = chromedp.Run(timeoutCtx, chromedp.Click(preSelectors, queryActions...)); err != nil {
							return errors.Wrap(err, fmt.Sprintf("event before Click Error: %s", preSelectors))
						}
					}
				case SetAttributesValueEvent:
					if flag {
						for _, setAttr := range event.Selector.SetAttrs {
							timeoutCtx, nodeCancel := context.WithTimeout(ctx, conf.Timeout)
							defer nodeCancel()
							if err = chromedp.Run(timeoutCtx, chromedp.SetAttributeValue(preSelectors, setAttr.AttributeName, setAttr.AttributeValue, queryActions...)); err != nil {
								return errors.Wrap(err, fmt.Sprintf("event before SetAttributesValue Error: %s", preSelectors))
							}
						}
					}
				default:
					logrus.Error(fmt.Sprintf("event type unmatched: %s", event.Type))
				}
			}
		}
	}
	return nil
}

// paginateCondition proceed except last page
func paginateCondition(ctx context.Context, conf config.RabiConfig, job Job, father *cdp.Node) (goon bool, err error) {
	var (
		queryActions []chromedp.QueryOption
	)
	queryActions = append(queryActions, chromedp.ByQuery)
	if father != nil {
		queryActions = append(queryActions, chromedp.FromNode(father))
	}
	return ExecEventCondition(ctx, conf, job.Paginator.Condition, queryActions)
}

func ExecEventCondition(ctx context.Context, conf config.RabiConfig, condition *Condition, queryActions []chromedp.QueryOption) (bool, error) {
	if condition == nil || stringutils.IsEmpty(condition.Value) {
		return true, nil
	}
	var (
		text string
		ok   bool
	)
	conditionTimeoutCtx, conditionCancel := context.WithTimeout(ctx, conf.Timeout)
	defer conditionCancel()
	switch condition.ExecSelector.Type {
	case TextEvent:
		conditionCss := CssOrXpath(condition.ExecSelector.Selector)
		if err := chromedp.Run(conditionTimeoutCtx, chromedp.Text(conditionCss, &text, queryActions...)); err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("condition css err: %s", conditionCss))
		}
		logrus.Infof("retrieve value: %s, expect: %s\n", text, condition.Value)
		return condition.CheckFunc(text, condition.Value), nil
	case GetAttributeValueEvent:
		execSelector := condition.ExecSelector.Selector
		if err := chromedp.Run(conditionTimeoutCtx, chromedp.AttributeValue(execSelector.Css, execSelector.Attr, &text, &ok, queryActions...)); err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("condition css: %s, attr: %s err", execSelector.Css, execSelector.Attr))
		}
		logrus.Infof("attribute: %s, retrieve value: %s, expect: %s\n", execSelector.Attr, text, condition.Value)
		return condition.CheckFunc(text, condition.Value), nil
	default:
		logrus.Error(fmt.Sprintf("event type unmatched: %s", condition.ExecSelector.Type))
	}
	return true, nil
}

func (r RabidaImpl) CrawlWithConfig(ctx context.Context, job Job, callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, conf config.RabiConfig, options ...chromedp.ExecAllocatorOption) error {
	return r.CrawlWithListeners(ctx, job, func(ctx context.Context, ret []interface{}, nextPageUrl string, currentPageNo int) bool {
		return callback(ret, nextPageUrl, currentPageNo)
	}, before, after, &conf, options)
}

func iframe(ctx context.Context, timeout time.Duration, job Job) (iframe *cdp.Node, err error) {
	var iframes []*cdp.Node
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if job.CssSelector.IframeSelector != nil {
		cssSelector := CssOrXpath(*job.CssSelector.IframeSelector)
		if err = chromedp.Run(timeoutCtx, chromedp.Nodes(cssSelector, &iframes, chromedp.BySearch)); err != nil {
			return nil, errors.Wrap(err, "iframe err")
		}
	} else {
		if err = chromedp.Run(timeoutCtx, chromedp.Nodes("iframe", &iframes, chromedp.ByQuery)); err != nil {
			return nil, errors.Wrap(err, "")
		}
	}
	if len(iframes) > 0 {
		iframe = iframes[0]
	}
	return
}

func (r RabidaImpl) sleep(conf config.RabiConfig) {
	DelaySleep(conf, "crawl next page")
}

func DelaySleep(conf config.RabiConfig, tag string) {
	var s time.Duration
	if len(conf.Delay) > 1 {
		s = lib.RandDuration(conf.Delay[0], conf.Delay[1])
	} else {
		s = conf.Delay[0]
	}
	logrus.Infof("\n delay sleep %s, tag: %s", s.String(), tag)
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
				logrus.Error(fmt.Sprintf("scope err: %+v", errors.Wrap(ErrNotFound, scope)))
			}
		} else {
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll)); err != nil {
				logrus.Error(fmt.Sprintf("scope err: %+v", errors.Wrap(ErrNotFound, scope)))
			}
		}
	} else {
		nodes = append(nodes, father)
	}
	var ret []interface{}
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

func (r RabidaImpl) extract(ctx context.Context, job Job, pageNo int, conf config.RabiConfig) (ret []interface{}, nextPageUrl string, err error) {
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

	p := r.paginator(job, pageNo)
	if stringutils.IsNotEmpty(job.CssSelector.XpathScope) || stringutils.IsNotEmpty(job.CssSelector.Xpath) {
		doc := r.Html(ctx, father, conf)
		ret = r.populateX(ctx, job.CssSelector, conf, doc)
		if stringutils.IsNotEmpty(p.Xpath) {
			nextPageUrl = lib.FindOne(doc, p.Xpath)
		}
	} else {
		ret = r.populate(ctx, father, job.CssSelector, conf)
		if stringutils.IsNotEmpty(p.Css) && stringutils.IsNotEmpty(p.Attr) {
			timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
			defer cancel()
			_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(p.Css, p.Attr, &nextPageUrl, chromedp.ByQuery))
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
