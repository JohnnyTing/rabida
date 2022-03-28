package useragent

import (
	_ "embed"
	"encoding/json"
)

const (
	CHROME            = "chrome"
	CHROME_MAC        = "chrome-mac"
	INTERNET_EXPLORER = "internet-explorer"
	FIREFOX           = "firefox"
	SAFARI            = "safari"

	ANDROID  = "android"
	MAC_OS_X = "mac-os-x"
	IOS      = "ios"
	LINUX    = "linux"

	IPHONE = "iphone"
	IPAD   = "ipad"

	COMPUTER = "computer"
	MOBILE   = "mobile"
)

//go:embed fake_useragent_0.2.0.json
var UaJsonBytes []byte
var UserAgent map[string][]string
var PcKeys []string

func init() {
	err := json.Unmarshal(UaJsonBytes, &UserAgent)
	if err != nil {
		panic(err)
	}
	PcKeys = append(PcKeys, CHROME, FIREFOX, SAFARI, MAC_OS_X, LINUX)
}
