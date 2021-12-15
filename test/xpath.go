package main

import (
	"github.com/sirupsen/logrus"
	"regexp"
)

func main() {
	xpath := `//div[@id="indexCarousel"]//div[@class="item"]//img/@src`
	reg := regexp.MustCompile(`/@(\w+$)`)
	if reg.MatchString(xpath) {
		rs := reg.FindString(xpath)
		logrus.Infoln(rs)
		submatch := reg.FindStringSubmatch(xpath)
		for _, result := range submatch {
			logrus.Infoln(result)
		}
		logrus.Infoln(submatch[1])

	}

}
