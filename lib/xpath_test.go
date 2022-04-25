package lib

import (
	"regexp"
	"testing"
)

func TestXpathAttr(t *testing.T) {
	tests := []struct {
		name  string
		xpath string
		want  string
	}{
		{
			name:  "",
			xpath: `//div[@id="indexCarousel"]//div[@class="item"]//img/@src`,
			want:  "src",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if is, attr := XpathAttr(tt.xpath); is && attr == tt.want {
				t.Logf("test passed, value: %s, want: %s", attr, tt.want)
			} else {
				t.Errorf("XpathAttr() = %v, want %v", attr, tt.want)

			}
		})
	}
}

func TestNthChildIndexXpath(t *testing.T) {
	data := []struct {
		xpath string
		want  int
	}{
		{
			"/html/body/div[4]/div/div[3]/div[3]/div[1]/div[60]",
			60,
		},
	}
	for _, item := range data {
		rs, err := NthChildFromXpath(item.xpath)
		if err != nil {
			t.Error(err)
		}
		t.Log(rs)
		if rs != item.want {
			t.Errorf("got: %v, want: %v", rs, item.want)
		}
	}
}

func TestNodeConditionFromXpath(t *testing.T) {
	data := []struct {
		xpath string
		want  string
	}{
		{
			"/bookstore/book[price>35.00]",
			"price>35.00",
		},
		{
			"/bookstore/book",
			"price>35.00",
		},
	}
	for _, item := range data {
		rs, exist := NodeConditionFromXpath(item.xpath)
		t.Log(rs)
		if exist && rs != item.want {
			t.Errorf("got: %s, want: %s", rs, item.want)
		}
	}
}

func TestBuildScopeByPosition(t *testing.T) {
	data := regexp.MustCompile(`\[(.+)\]$`).ReplaceAllString("/bookstore/book[price>35.00]", "")
	t.Log(data)

	scope := CursorScopeByPosition("/bookstore/book[price>35.00]", 15)
	t.Log(scope)
}
