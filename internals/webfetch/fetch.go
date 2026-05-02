package webfetch

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// WebFetcher simulates fetching a URL with a random delay, then sends the result
// to the provided channel. The channel is closed once the result is sent.
func WebFetcher(url string, ch chan string) {
	// Ensure the channel is closed when the function exits,
	// signaling to receivers that no more values will be sent.
	defer close(ch)

	// Simulate variable network latency with a random delay up to 100ms.
	delay := time.Duration(rand.IntN(100)) * time.Millisecond
	time.Sleep(delay)

	// Send the fetched result as a formatted string into the channel.
	ch <- fmt.Sprintf("Fetched: %s", url)
}
