package useragent

import (
	"math/rand"
	"time"
)

func RandomPcUA() string {
	rand.Seed(time.Now().UnixNano())
	keyIndex := rand.Intn(len(PcKeys))
	key := PcKeys[keyIndex]
	list := UserAgent[key]
	listIndex := rand.Intn(len(list))
	return list[listIndex]
}

func RandomMacChromeUA() string {
	rand.Seed(time.Now().UnixNano())
	list := UserAgent[CHROME_MAC]
	listIndex := rand.Intn(len(list))
	return list[listIndex]
}
