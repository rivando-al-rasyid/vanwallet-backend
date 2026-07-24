package service

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type MidtransService struct {
	cfg    config.MidtransConfig
	client *http.Client
}

func NewMidtransService(cfg config.MidtransConfig) *MidtransService {
	return &MidtransService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type SnapTransactionResult struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type snapCreateRequest struct {
	TransactionDetails snapTransactionDetails `json:"transaction_details"`
	EnabledPayments    []string             `json:"enabled_payments"`
	CustomerDetails    *snapCustomerDetails `json:"customer_details,omitempty"`
	Callbacks          *snapCallbacks       `json:"callbacks,omitempty"`
}

type snapTransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type snapCustomerDetails struct {
	Email string `json:"email,omitempty"`
}

type snapCallbacks struct {
	Finish string `json:"finish,omitempty"`
}

func (m *MidtransService) ClientKey() string {
	return m.cfg.ClientKey
}

func (m *MidtransService) CreateSnapTransaction(orderID string, amount int64, paymentMethod model.PaymentMethod, customerEmail string) (SnapTransactionResult, error) {
	enabledPayment, ok := paymentMethodToMidtrans(paymentMethod)
	if !ok {
		return SnapTransactionResult{}, fmt.Errorf("unsupported payment method for Midtrans: %s", paymentMethod)
	}

	reqBody := snapCreateRequest{
		TransactionDetails: snapTransactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
		EnabledPayments: []string{enabledPayment},
	}

	if customerEmail != "" {
		reqBody.CustomerDetails = &snapCustomerDetails{Email: customerEmail}
	}
	if m.cfg.FinishURL != "" {
		reqBody.Callbacks = &snapCallbacks{Finish: m.cfg.FinishURL}
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return SnapTransactionResult{}, fmt.Errorf("marshal snap request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, m.cfg.SnapBaseURL()+"/snap/v1/transactions", bytes.NewReader(payload))
	if err != nil {
		return SnapTransactionResult{}, fmt.Errorf("create snap request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(m.cfg.ServerKey, "")

	resp, err := m.client.Do(req)
	if err != nil {
		return SnapTransactionResult{}, fmt.Errorf("call snap api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SnapTransactionResult{}, fmt.Errorf("read snap response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SnapTransactionResult{}, fmt.Errorf("snap api returned %d: %s", resp.StatusCode, string(body))
	}

	var result SnapTransactionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return SnapTransactionResult{}, fmt.Errorf("decode snap response: %w", err)
	}
	if result.Token == "" {
		return SnapTransactionResult{}, fmt.Errorf("snap api returned empty token")
	}

	return result, nil
}

func (m *MidtransService) VerifyNotificationSignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	if signatureKey == "" {
		return false
	}

	raw := orderID + statusCode + grossAmount + m.cfg.ServerKey
	sum := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(sum[:])
	return strings.EqualFold(expected, signatureKey)
}

func paymentMethodToMidtrans(pm model.PaymentMethod) (string, bool) {
	switch pm {
	case model.PaymentMethodBRI:
		return "bri_va", true
	case model.PaymentMethodBCA:
		return "bca_va", true
	case model.PaymentMethodDANA:
		return "dana", true
	case model.PaymentMethodGoPay:
		return "gopay", true
	case model.PaymentMethodOVO:
		return "ovo", true
	default:
		return "", false
	}
}

func MidtransStatusToTopupStatus(transactionStatus, fraudStatus string) (model.TransactionStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(transactionStatus)) {
	case "capture":
		if strings.EqualFold(fraudStatus, "accept") || fraudStatus == "" {
			return model.TransactionStatusSuccess, true
		}
		return model.TransactionStatusFailed, true
	case "settlement":
		return model.TransactionStatusSuccess, true
	case "deny", "failure":
		return model.TransactionStatusFailed, true
	case "cancel":
		return model.TransactionStatusCancelled, true
	case "expire":
		return model.TransactionStatusFailed, true
	default:
		return "", false
	}
}
