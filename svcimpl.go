package service

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"github.com/unionj-cloud/rabida/config"
	"github.com/unionj-cloud/rabida/lib"
	"time"
)

type RabidaImpl struct {
	conf *config.RabiConfig
}

func (r RabidaImpl) Crawl(ctx context.Context, job Job, callback func(ret []map[string]string, nextPageUrl string, currentPageNo int) bool) error {
	var (
		err         error
		abort       bool
		ret         []map[string]string
		nextPageUrl string
		pageNo      int
		cancel      context.CancelFunc
	)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"),
		chromedp.Headless,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	}

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	if err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			_, err = page.AddScriptToEvaluateOnNewDocument(lib.Script).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		chromedp.Navigate(job.Link),
	}); err != nil {
		return errors.Wrap(err, "")
	}

	ret, nextPageUrl, err = r.extract(ctx, job)
	if err != nil {
		return errors.Wrap(err, "")
	}

	pageNo++
	if abort = callback(ret, nextPageUrl, pageNo); abort {
		return nil
	}

	if stringutils.IsEmpty(job.Paginator) {
		return nil
	}

	fmt.Printf("sleep 1s to crawl next page\n")
	time.Sleep(r.conf.Delay[0])

	for {
		var (
			timeoutCtx context.Context
			cancel     context.CancelFunc
		)
		timeoutCtx, cancel = context.WithTimeout(ctx, r.conf.Timeout)
		if err = chromedp.Run(timeoutCtx, chromedp.Click(job.Paginator, chromedp.ByQuery)); err != nil {
			goto ERR
		}
		time.Sleep(r.conf.Timeout)
		if ret, nextPageUrl, err = r.extract(ctx, job); err != nil {
			goto ERR
		}
		pageNo++
		if abort = callback(ret, nextPageUrl, pageNo); abort {
			goto END
		}
		cancel()
		fmt.Printf("sleep 1s to crawl next page\n")
		time.Sleep(r.conf.Delay[0])
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

type errNotFound struct{}

func (e errNotFound) Error() string {
	return "not found"
}

var ErrNotFound error = errNotFound{}

func (r RabidaImpl) extract(ctx context.Context, job Job) ([]map[string]string, string, error) {
	var (
		err error
		ret []map[string]string
	)

	if stringutils.IsEmpty(job.Scope) {
		if ret, err = r.noscope(ctx, job); err != nil {
			return nil, "", errors.Wrap(err, "")
		}
	} else {
		if ret, err = r.scope(ctx, job); err != nil {
			return nil, "", errors.Wrap(err, "")
		}
	}

	var nextPageUrl string
	var ok bool
	timeoutCtx, cancel := context.WithTimeout(ctx, r.conf.Timeout)
	defer cancel()
	_ = chromedp.Run(timeoutCtx, chromedp.AttributeValue(job.Paginator, "href", &nextPageUrl, &ok,
		chromedp.ByQuery))
	return ret, nextPageUrl, nil
}

func (r RabidaImpl) scope(ctx context.Context, job Job) ([]map[string]string, error) {
	var (
		nodes []*cdp.Node
		err   error
		ret   []map[string]string
	)
	timeoutCtx, cancel := context.WithTimeout(ctx, r.conf.Timeout)
	defer cancel()
	if err = chromedp.Run(timeoutCtx, chromedp.Nodes(job.Scope, &nodes)); err != nil {
		return nil, errors.Wrap(ErrNotFound, "")
	}
	for _, node := range nodes {
		data := make(map[string]string)
		for attr, css := range job.Attrs {
			timeoutCtx, cancel = context.WithTimeout(ctx, r.conf.Timeout)
			var value string
			if stringutils.IsEmpty(css.Attr) {
				if err = chromedp.Run(timeoutCtx, chromedp.Text(css.Css, &value, chromedp.ByQuery, chromedp.FromNode(node))); err != nil {
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

func (r RabidaImpl) noscope(ctx context.Context, job Job) ([]map[string]string, error) {
	var (
		err error
		ret []map[string]string
	)
	data := make(map[string]string)
	for attr, css := range job.Attrs {
		timeoutCtx, cancel := context.WithTimeout(ctx, r.conf.Timeout)
		var value string
		if stringutils.IsEmpty(css.Attr) {
			if err = chromedp.Run(timeoutCtx, chromedp.Text(css.Css, &value, chromedp.ByQuery)); err != nil {
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
