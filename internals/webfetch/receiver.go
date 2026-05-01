package webfetch

import (
	"log"
)

func Receiver(ch chan string) {
	for result := range ch {
		log.Println(result)
	}
}
