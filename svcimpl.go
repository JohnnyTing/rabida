package service

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"github.com/unionj-cloud/rabida/config"
	"github.com/unionj-cloud/rabida/lib"
	"time"
)

type RabidaImpl struct {
	conf *config.RabiConfig
}

func (r RabidaImpl) CrawlWithConfig(ctx context.Context, job Job, callback func(ret []map[string]string, nextPageUrl string, currentPageNo int) bool, before []chromedp.Action, after []chromedp.Action, conf config.RabiConfig) error {
	var (
		err         error
		abort       bool
		ret         []map[string]string
		nextPageUrl string
		pageNo      int
		cancel      context.CancelFunc
		timeoutCtx  context.Context
	)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	}
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

func (r RabidaImpl) Crawl(ctx context.Context, job Job, callback func(ret []map[string]string, nextPageUrl string, currentPageNo int) bool,
	before []chromedp.Action, after []chromedp.Action) error {
	return r.CrawlWithConfig(ctx, job, callback, before, after, *r.conf)
}

type errNotFound struct{}

func (e errNotFound) Error() string {
	return "not found"
}

var ErrNotFound error = errNotFound{}

func (r RabidaImpl) extract(ctx context.Context, job Job, conf config.RabiConfig) ([]map[string]string, string, error) {
	var (
		err error
		ret []map[string]string
	)

	if stringutils.IsEmpty(job.Scope) {
		if ret, err = r.noscope(ctx, job, conf); err != nil {
			return nil, "", errors.Wrap(err, "")
		}
	} else {
		if ret, err = r.scope(ctx, job, conf); err != nil {
			return nil, "", errors.Wrap(err, "")
		}
	}

	var nextPageUrl string
	var ok bool
	timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	_ = chromedp.Run(timeoutCtx, chromedp.AttributeValue(job.Paginator.Css, job.Paginator.Attr, &nextPageUrl, &ok,
		chromedp.ByQuery))
	return ret, nextPageUrl, nil
}

func (r RabidaImpl) scope(ctx context.Context, job Job, conf config.RabiConfig) ([]map[string]string, error) {
	var (
		nodes []*cdp.Node
		err   error
		ret   []map[string]string
	)
	timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	if err = chromedp.Run(timeoutCtx, chromedp.Nodes(job.Scope, &nodes)); err != nil {
		return nil, errors.Wrap(ErrNotFound, "")
	}
	for _, node := range nodes {
		data := make(map[string]string)
		for attr, css := range job.Attrs {
			timeoutCtx, cancel = context.WithTimeout(ctx, conf.Timeout)
			var value string
			if stringutils.IsEmpty(css.Attr) {
				if err = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(css.Css, "innerText", &value, chromedp.ByQuery, chromedp.FromNode(node))); err != nil {
					goto ERR
				}
			} else {
				var ok bool
				if err = chromedp.Run(timeoutCtx, chromedp.AttributeValue(css.Css, css.Attr, &value, &ok, chromedp.ByQuery, chromedp.FromNode(node))); err != nil {
					goto ERR
				}
			}
			data[attr] = value
		ERR:
			cancel()
		}
		ret = append(ret, data)
	}
	return ret, nil
}

func (r RabidaImpl) noscope(ctx context.Context, job Job, conf config.RabiConfig) ([]map[string]string, error) {
	var (
		err error
		ret []map[string]string
	)
	data := make(map[string]string)
	for attr, css := range job.Attrs {
		timeoutCtx, cancel := context.WithTimeout(ctx, conf.Timeout)
		var value string
		if stringutils.IsEmpty(css.Attr) {
			if err = chromedp.Run(timeoutCtx, chromedp.JavascriptAttribute(css.Css, "innerText", &value, chromedp.ByQuery)); err != nil {
				goto ERR
			}
		} else {
			var ok bool
			if err = chromedp.Run(timeoutCtx, chromedp.AttributeValue(css.Css, css.Attr, &value, &ok, chromedp.ByQuery)); err != nil {
				goto ERR
			}
		}
		data[attr] = value
	ERR:
		cancel()
	}
	ret = append(ret, data)
	return ret, nil
}

func NewRabida(conf *config.RabiConfig) Rabida {
	return &RabidaImpl{
		conf,
	}
}
