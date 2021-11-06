package service

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"github.com/unionj-cloud/rabida/config"
	"github.com/unionj-cloud/rabida/internal/lib"
	"time"
)

type RabidaImpl struct {
	conf *config.RabiConfig
}

func (r RabidaImpl) CrawlWithConfig(ctx context.Context, job Job, callback func(ret []interface{}, nextPageUrl string, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, conf config.RabiConfig, options ...chromedp.ExecAllocatorOption) error {
	var (
		err         error
		abort       bool
		ret         []interface{}
		nextPageUrl string
		pageNo      int
		cancel      context.CancelFunc
		timeoutCtx  context.Context
	)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
	}
	opts = append(opts, options...)
	if conf.Mode == "headless" {
		opts = append(opts, chromedp.Headless)
	}

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

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
	tasks = append(tasks, chromedp.Navigate(link))
	tasks = append(tasks, after...)

	timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
	defer cancel()

	if err = chromedp.Run(timeoutCtx, tasks); err != nil {
		return errors.Wrap(err, "")
	}

	time.Sleep(conf.Timeout)

	if stringutils.IsNotEmpty(job.StartPageBtn.Css) {
		timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
		defer cancel()
		if err = chromedp.Run(timeoutCtx, chromedp.Click(job.StartPageBtn.Css, chromedp.ByQuery)); err != nil {
			return errors.Wrap(err, "")
		}
		time.Sleep(conf.Timeout)
	}

	ret, nextPageUrl, err = r.extract(ctx, job, conf)
	if err != nil {
		return errors.Wrap(err, "")
	}

	pageNo++
	if abort = callback(ret, nextPageUrl, pageNo); abort {
		return nil
	}

	if stringutils.IsEmpty(job.Paginator.Css) {
		return nil
	}

	r.sleep(conf)

	for {
		timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
		if err = chromedp.Run(timeoutCtx, chromedp.Click(job.Paginator.Css, chromedp.ByQuery)); err != nil {
			goto ERR
		}
		time.Sleep(conf.Timeout)
		if ret, nextPageUrl, err = r.extract(ctx, job, conf); err != nil {
			goto ERR
		}
		pageNo++
		if abort = callback(ret, nextPageUrl, pageNo); abort {
			goto END
		}
		cancel()

		r.sleep(conf)
		continue

	END:
		cancel()
		return nil
	ERR:
		cancel()
		if errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return errors.Wrap(err, "")
	}
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

func (r RabidaImpl) populate(ctx context.Context, scope string, father *cdp.Node, cssSelector CssSelector, conf config.RabiConfig) []interface{} {
	var nodes []*cdp.Node
	timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	if stringutils.IsNotEmpty(scope) {
		if father != nil {
			timeoutCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll, chromedp.FromNode(father))); err != nil {
				var html string
				err = chromedp.Run(ctx, chromedp.OuterHTML("html", &html, chromedp.ByQuery))
				if err != nil {
					panic(errors.Wrap(ErrNotFound, ""))
				}
				fmt.Println(html)
				panic(errors.Wrap(ErrNotFound, ""))
			}
		} else {
			if err := chromedp.Run(timeoutCtx, chromedp.Nodes(scope, &nodes, chromedp.ByQueryAll)); err != nil {
				panic(errors.Wrap(ErrNotFound, ""))
			}
		}
	} else {
		nodes = append(nodes, father)
	}
	var ret []interface{}
	for _, node := range nodes {
		if cssSelector.Attrs == nil {
			timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
			var value string
			if stringutils.IsEmpty(cssSelector.Attr) {
				if cssSelector.Css == ":scope" {
					_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute([]cdp.NodeID{node.NodeID}, "innerText", &value, chromedp.ByNodeID))
				} else {
					_ = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(cssSelector.Css, "innerText", &value, chromedp.ByQuery, chromedp.FromNode(node)))
				}
			} else {
				var ok bool
				if cssSelector.Css == ":scope" {
					if cssSelector.Attr == "outerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.OuterHTML([]cdp.NodeID{node.NodeID}, &value, chromedp.ByNodeID))
					} else if cssSelector.Attr == "innerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.InnerHTML([]cdp.NodeID{node.NodeID}, &value, chromedp.ByNodeID))
					} else {
						_ = chromedp.Run(timeoutCtx, chromedp.AttributeValue([]cdp.NodeID{node.NodeID}, cssSelector.Attr, &value, &ok, chromedp.ByNodeID))
					}
				} else {
					if cssSelector.Attr == "outerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.OuterHTML(cssSelector.Css, &value, chromedp.ByQuery, chromedp.FromNode(node)))
					} else if cssSelector.Attr == "innerHTML" {
						_ = chromedp.Run(timeoutCtx, chromedp.InnerHTML(cssSelector.Css, &value, chromedp.ByQuery, chromedp.FromNode(node)))
					} else {
						_ = chromedp.Run(timeoutCtx, chromedp.AttributeValue(cssSelector.Css, cssSelector.Attr, &value, &ok, chromedp.ByQuery, chromedp.FromNode(node)))
					}
				}
			}
			if stringutils.IsNotEmpty(value) {
				ret = append(ret, value)
			}
			cancel()
		} else {
			data := make(map[string]interface{})
			for attr, sel := range cssSelector.Attrs {
				result := r.populate(ctx, sel.Scope, node, sel, conf)
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

	rootScope := job.Scope
	if stringutils.IsEmpty(rootScope) {
		rootScope = "html"
	}
	ret = r.populate(ctx, rootScope, nil, job.CssSelector, conf)
	var ok bool
	timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	_ = chromedp.Run(timeoutCtx, chromedp.AttributeValue(job.Paginator.Css, job.Paginator.Attr, &nextPageUrl, &ok,
		chromedp.ByQuery))
	return
}

func NewRabida(conf *config.RabiConfig) Rabida {
	return &RabidaImpl{
		conf,
	}
}
