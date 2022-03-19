package useragent

import (
	"encoding/json"
	"github.com/unionj-cloud/go-doudou/toolkit/pathutils"
	"io/ioutil"
)

const (
	CHROME            = "chrome"
	INTERNET_EXPLORER = "internet-explorer"
	FIREFOX           = "firefox"
	SAFARI            = "safari"

	ANDROID  = "android"
	MAC_OS_X = "mac-os-x"
	IOS      = "ios"
	LINUX    = "linux"

	IPHONE = "iphone"
	IPAD   = "ipad"

	COMPUTER   = "computer"
	MOBILE     = "mobile"
	UaJsonPath = "fake_useragent_0.2.0.json"
)

var UserAgent map[string][]string
var PcKeys []string

func init() {
	path := pathutils.Abs(UaJsonPath)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &UserAgent)
	if err != nil {
		panic(err)
	}
	PcKeys = append(PcKeys, CHROME, FIREFOX, SAFARI, MAC_OS_X, LINUX)
}
