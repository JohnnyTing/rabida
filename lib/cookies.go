package lib

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"net/http"
	"time"
)

func HttpCookies(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{Header: header}
	return request.Cookies()
}

func CookieAction(rawCookies, domain string, expire int) chromedp.ActionFunc {
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
		//rawCookies := "iibulanguage=CN; ibulocale=zh_cn; cookiePricesDisplayed=CNY; _bfaStatusPVSend=1; _RSG=4Xai51hp.5FX0dOYkEpn7B; _RDG=289e3e7b65b1772261218d2c5cbb85be06; _RGUID=f17eceb0-45b2-4844-b1f1-cceba2627021; MKT_CKID=1638145516014.jy1ma.beis; GUID=09031121319994645452; nfes_isSupportWebP=1; login_uid=754BF301C3EBA5CD980F210BDADB599E; login_type=0; AHeadUserInfo=VipGrade=0&VipGradeName=%C6%D5%CD%A8%BB%E1%D4%B1&UserName=&NoReadMessageCount=0; UUID=247FFBC37C6248E39753176CBBB8E98E; IsPersonalizedLogin=T; __zpspc=9.29.1639731599.1639731652.2%234%7C%7C%7C%7C%7C%23; _jzqco=%7C%7C%7C%7C%7C1.2121005470.1638145516018.1639731599800.1639731652458.1639731599800.1639731652458.0.0.0.171.171; appFloatCnt=64; intl_ht1=h4=30_635264,22033_4831432,22033_1687298,22033_1687326,22033_6745255,22033_1687361; _RF1=218.88.126.107; cticket=BF79374B1282A6DEC82AF3355D15DEB1C9C487C019093728E5576511B8ED5FC5; DUID=u=754BF301C3EBA5CD980F210BDADB599E&v=0; IsNonUser=F; librauuid=3hZ62Z4QUgbUV9XF; _bfa=1.1638145450811.242vv1.1.1639731595891.1640765767392.31.196; _bfs=1.3; _bfaStatus=send"
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
