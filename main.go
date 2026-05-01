package main

import (
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/webfetch"
)

func main() {

	urls := []string{
		"https://example.com",
		"https://golang.org",
		"https://github.com",
		"https://openai.com",
		"https://anthropic.com",
	}

	for _, url := range urls {
		urlch := make(chan string)

		go webfetch.WebFetcher(url, urlch)
		webfetch.Receiver(urlch)
	}

}
