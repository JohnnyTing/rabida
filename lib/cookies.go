package lib

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

func HttpCookies(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{Header: header}
	return request.Cookies()
}

func CookieAction(link, rawCookies string, expire int) chromedp.ActionFunc {
	rs, _ := url.Parse(link)
	domain := rs.Hostname()
	logrus.Println(domain)
	return func(ctx context.Context) (err error) {
		// create cookie expiration
		var duration time.Duration
		if expire == 0 {
			duration = time.Duration(360 * 24)
		} else {
			duration = time.Duration(expire)
		}
		expr := cdp.TimeSinceEpoch(time.Now().Add(duration * time.Hour))
		// add cookies to chrome
		cookies := HttpCookies(rawCookies)
		for _, item := range cookies {
			// 设置cookies
			err := network.SetCookie(item.Name, item.Value).
				WithExpires(&expr).
				WithDomain(domain).
				Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
