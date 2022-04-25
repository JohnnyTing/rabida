package lib

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"golang.org/x/net/html"
	"regexp"
)

func FindOne(node *html.Node, xpath string) string {
	nodeOne, err := htmlquery.Query(node, xpath)
	if err != nil {
		panic("can't reconized xpath expression")
	}
	is, attr := XpathAttr(xpath)
	if is {
		return htmlquery.SelectAttr(nodeOne, attr)
	}
	return htmlquery.InnerText(nodeOne)
}

func XpathAttr(xpath string) (flag bool, attr string) {
	flag = false
	reg := regexp.MustCompile(`/@(\w+$)`)
	if reg.MatchString(xpath) {
		flag = true
		submatch := reg.FindStringSubmatch(xpath)
		attr = submatch[1]
	}
	return
}

func NthChildFromXpath(xpath string) (cursor int, err error) {
	reg := regexp.MustCompile(`\[(\d+)\]$`)
	if reg.MatchString(xpath) {
		submatch := reg.FindStringSubmatch(xpath)
		return cast.ToInt(submatch[1]), err
	}
	return 0, errors.New("not find xpath nth-child index")
}

func NodeConditionFromXpath(xpath string) (condition string, exist bool) {
	reg := regexp.MustCompile(`\[(.+)\]$`)
	if reg.MatchString(xpath) {
		submatch := reg.FindStringSubmatch(xpath)
		return submatch[1], true
	}
	return "", false
}

func CursorScopeByPosition(xpathScope string, cursor int) (scope string) {
	condition, exist := NodeConditionFromXpath(xpathScope)
	if exist {
		prefix := regexp.MustCompile(`\[(.+)\]$`).ReplaceAllString(xpathScope, "")
		scope = fmt.Sprintf("%s[%s and %v<=position()]", prefix, condition, cursor)
	} else {
		scope = fmt.Sprintf("%s[%v<=position()]", xpathScope, cursor)
	}
	return scope
}
