package config

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SMTPConfig() (*gomail.Dialer, error) {
	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		return nil, err
	}

	d := gomail.NewDialer(
		os.Getenv("EMAIL_HOST"),
		port,
		os.Getenv("EMAIL_USERNAME"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	return d, nil
}
