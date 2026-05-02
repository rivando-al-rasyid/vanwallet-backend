package main

import (
	"fmt"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/user"
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

	um := user.NewUserManager()

	fmt.Println("=== Manajemen User ===")

	fmt.Println("\n-- test AddUser --")
	if err := um.AddUser("001", "Budi Santoso"); err != nil {
		fmt.Println(err)
	}
	if err := um.AddUser("002", "Siti Rahayu"); err != nil {
		fmt.Println(err)
	}
	if err := um.AddUser("003", "Andi Wijaya"); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n-- test Duplicate --")
	if err := um.AddUser("002", "Budi Lagi"); err != nil {
		fmt.Println(err)
	}

}
