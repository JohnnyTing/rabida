package lib

import (
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
