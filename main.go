package main

import (
	"sync"
)

func main() {
	var wg sync.WaitGroup
	urls := []string{
		"https://example.com",
		"https://golang.org",
		"https://github.com",
		"https://openai.com",
		"https://anthropic.com",
	}

	urlchan := make(chan string)

	// Fetcher goroutines

	// Receiver goroutine — delegates to receiver.go

	wg.Wait()
}
