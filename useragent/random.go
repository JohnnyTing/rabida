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

func RandomMacUA() string {
	rand.Seed(time.Now().UnixNano())
	list := UserAgent[MAC_OS_X]
	listIndex := rand.Intn(len(list))
	return list[listIndex]
}
