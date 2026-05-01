package webfetch

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func WebFetcher(url string, ch chan string) {
	delay := time.Duration(rand.IntN(100)) * time.Millisecond
	time.Sleep(delay)
	ch <- fmt.Sprintf("Url: %s", url)
}
