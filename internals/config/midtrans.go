package config

import (
	"fmt"
	"os"
	"strings"
)

type MidtransConfig struct {
	ServerKey   string
	ClientKey   string
	IsProduction bool
	FinishURL   string
}

func MidtransConfigFromEnv() (MidtransConfig, error) {
	serverKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	if serverKey == "" {
		return MidtransConfig{}, fmt.Errorf("MIDTRANS_SERVER_KEY is required")
	}

	clientKey := strings.TrimSpace(os.Getenv("MIDTRANS_CLIENT_KEY"))
	if clientKey == "" {
		return MidtransConfig{}, fmt.Errorf("MIDTRANS_CLIENT_KEY is required")
	}

	isProduction := strings.EqualFold(strings.TrimSpace(os.Getenv("MIDTRANS_IS_PRODUCTION")), "true")

	return MidtransConfig{
		ServerKey:    serverKey,
		ClientKey:    clientKey,
		IsProduction: isProduction,
		FinishURL:    strings.TrimSpace(os.Getenv("MIDTRANS_FINISH_URL")),
	}, nil
}

func (c MidtransConfig) SnapBaseURL() string {
	if c.IsProduction {
		return "https://app.midtrans.com"
	}
	return "https://app.sandbox.midtrans.com"
}
