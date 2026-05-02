package webfetch

import (
	"log"
)

// Receiver reads fetch results from the channel and logs each one.
// It blocks until the channel is closed, processing results as they arrive.
func Receiver(ch chan string) {
	// Iterate over incoming messages from the channel.
	// The loop exits automatically when the channel is closed.
	for result := range ch {
		log.Println(result)
	}
}
