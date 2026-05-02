package main

import (
	"fmt"

	calculator "github.com/rivando-al-rasyid/vanwallet-backend/internals/area"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/process"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/user"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/webfetch"
)

func main() {

	// ─────────────────────────────────────────────────────────────────────────────
	// 1. ProcessNumber
	//    Menguji pemrosesan slice integer dengan tiga skenario:
	//    input valid, input nil, dan input kosong (memicu panic yang di-recover).
	// ─────────────────────────────────────────────────────────────────────────────
	fmt.Println("\n=== ProcessNumber ===")

	fmt.Println("\n-- input valid --")
	if err := process.ProcessNumber([]int{1, 2, 3, 4, 5}); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("\n-- input nil --")
	if err := process.ProcessNumber(nil); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("\n-- input slice kosong (memicu panic/recover) --")
	if err := process.ProcessNumber([]int{}); err != nil {
		fmt.Println("Error:", err)
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// 2. WebFetch
	//    Mensimulasikan pengambilan data dari beberapa URL secara goroutine.
	//    Setiap URL diproses oleh WebFetcher dan hasilnya diterima via channel.
	//    Delay acak (rand.IntN) digunakan untuk mensimulasikan latensi jaringan.
	// ─────────────────────────────────────────────────────────────────────────────
	fmt.Println("\n=== WebFetch ===")

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

	// ─────────────────────────────────────────────────────────────────────────────
	// 3. UserManager
	//    Mensimulasikan manajemen pengguna: tambah user baru, deteksi duplikat ID,
	//    dan pencarian user berdasarkan ID (ditemukan & tidak ditemukan).
	// ─────────────────────────────────────────────────────────────────────────────
	fmt.Println("\n=== Manajemen User ===")

	um := user.NewUserManager()

	fmt.Println("\n-- menambahkan user --")
	if err := um.AddUser("001", "Budi Santoso"); err != nil {
		fmt.Println(err)
	}
	if err := um.AddUser("002", "Siti Rahayu"); err != nil {
		fmt.Println(err)
	}
	if err := um.AddUser("003", "Andi Wijaya"); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n-- menambahkan user dengan ID duplikat --")
	if err := um.AddUser("002", "Budi Lagi"); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n-- mencari user berdasarkan ID --")
	if u, err := um.GetUser("001"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("User ditemukan: Id=%s, Nama=%s\n", u.Id, u.Name)
	}

	fmt.Println("\n-- mencari user dengan ID yang tidak terdaftar --")
	if _, err := um.GetUser("999"); err != nil {
		fmt.Println(err)
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// 4. Area Calculator
	//    Menghitung luas Circle dan Rectangle menggunakan interface Shape.
	//    Semua shape dikumpulkan ke dalam slice, lalu total luas dihitung
	//    menggunakan CalculateTotalArea yang memanfaatkan channel secara internal.
	// ─────────────────────────────────────────────────────────────────────────────
	fmt.Println("\n=== Area Calculator ===")

	circle := calculator.Circle{Radius: 7}
	rect := calculator.Rectangle{Height: 4, Width: 6}

	fmt.Printf("Luas Circle    (r=7) : %.4f\n", circle.Area())
	fmt.Printf("Luas Rectangle (4x6) : %.4f\n", rect.Area())

	shapes := []calculator.Shape{circle, rect}
	total := calculator.CalculateTotalArea(shapes)
	fmt.Printf("Total luas           : %.4f\n", total)
}
