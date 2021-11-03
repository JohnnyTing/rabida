package service

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
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
		chromedp.ExecPath(r.conf.ChromePath),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
	}

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err = chromedp.Run(ctx, lib.Navigate(job.Link)); err != nil {
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

	fmt.Printf("sleep 1s to crawl next page\n")
	time.Sleep(r.conf.Delay[0])

	for {
		if err = chromedp.Run(ctx, chromedp.Click(job.Paginator, chromedp.ByQuery)); err != nil {
			return errors.Wrap(err, "")
		}
		time.Sleep(1 * time.Second)
		ret, nextPageUrl, err = r.extract(ctx, job)
		if err != nil {
			return errors.Wrap(err, "")
		}
		pageNo++
		if abort = callback(ret, nextPageUrl, pageNo); abort {
			return nil
		}
		fmt.Printf("sleep 1s to crawl next page\n")
		time.Sleep(r.conf.Delay[0])
	}
}

func (r RabidaImpl) extract(ctx context.Context, job Job) ([]map[string]string, string, error) {
	var (
		nodes []*cdp.Node
		err   error
	)
	timeoutCtx, cancel := context.WithTimeout(ctx, r.conf.Timeout)
	defer cancel()

	if err = chromedp.Run(timeoutCtx, chromedp.Nodes(job.Scope, &nodes)); err != nil {
		return nil, "", errors.Wrap(err, "")
	}

	var ret []map[string]string
	for _, node := range nodes {
		data := make(map[string]string)
		for attr, css := range job.Attrs {
			var value string
			if stringutils.IsEmpty(css.Attr) {
				if err = chromedp.Run(ctx, chromedp.Text(css.Css, &value, chromedp.ByQuery, chromedp.FromNode(node))); err != nil {
					return nil, "", errors.Wrap(err, "")
				}
			} else {
				var ok bool
				if err = chromedp.Run(ctx, chromedp.AttributeValue(css.Css, css.Attr, &value, &ok, chromedp.ByQuery, chromedp.FromNode(node))); err != nil {
					return nil, "", errors.Wrap(err, "")
				}
			}
			data[attr] = value
		}
		ret = append(ret, data)
	}

	var nextPageUrl string
	var ok bool
	if err = chromedp.Run(ctx, chromedp.AttributeValue(job.Paginator, "href", &nextPageUrl, &ok,
		chromedp.ByQuery)); err != nil {
		return nil, "", errors.Wrap(err, "")
	}
	return ret, nextPageUrl, nil
}

func NewRabida(conf *config.RabiConfig) Rabida {
	return &RabidaImpl{
		conf,
	}
}
