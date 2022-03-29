package lib

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/toolkit/stringutils"
)

func QueryIP(ip string, lang string) bool {
	client := resty.New().SetRetryCount(3)
	if stringutils.IsEmpty(lang) {
		lang = "zh-CN"
	}
	if stringutils.IsEmpty(ip) {
		logrus.Error("ip is empty")
		return false
	}
	//http://ip-api.com/json/120.220.220.95?lang=zh-CN
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	var rs Result
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"lang": lang,
		}).
		SetResult(&rs).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		logrus.Error(err)
		return false
	}
	if resp.IsSuccess() {
		logrus.Info(gabs.Wrap(rs).StringIndent("", "  "))
	} else {
		logrus.Error(resp.Error())
		return false
	}
	return true
}

type Result struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
}
