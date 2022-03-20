package useragent

import (
	"log"
	"testing"
)

func TestRandom_UA(t *testing.T) {
	for i := 0; i < 10; i++ {
		ua := RandomPcUA()
		log.Println(ua)
	}
}
