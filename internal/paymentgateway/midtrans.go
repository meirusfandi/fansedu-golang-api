package paymentgateway

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type MidtransClient struct {
	serverKey string
	baseURL   string
	httpc     *http.Client
}

func NewMidtransClient(serverKey string, isProduction bool, baseURLOverride string) *MidtransClient {
	baseURL := strings.TrimSpace(baseURLOverride)
	if baseURL == "" {
		if isProduction {
			baseURL = "https://app.midtrans.com"
		} else {
			baseURL = "https://app.sandbox.midtrans.com"
		}
	}
	return &MidtransClient{
		serverKey: strings.TrimSpace(serverKey),
		baseURL:   strings.TrimRight(baseURL, "/"),
		httpc:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *MidtransClient) Enabled() bool {
	return c != nil && c.serverKey != ""
}

type CreateSnapRequest struct {
	OrderID  string
	Amount   int
	Customer struct {
		FirstName string
		Email     string
	}
}

type CreateSnapResponse struct {
	Token       string
	RedirectURL string
}

func (c *MidtransClient) CreateSnapTransaction(ctx context.Context, req CreateSnapRequest) (CreateSnapResponse, error) {
	if !c.Enabled() {
		return CreateSnapResponse{}, errors.New("midtrans is disabled")
	}
	body := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     req.OrderID,
			"gross_amount": req.Amount,
		},
		"customer_details": map[string]any{
			"first_name": strings.TrimSpace(req.Customer.FirstName),
			"email":      strings.TrimSpace(req.Customer.Email),
		},
	}
	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/snap/v1/transactions", bytes.NewReader(b))
	if err != nil {
		return CreateSnapResponse{}, err
	}
	httpReq.SetBasicAuth(c.serverKey, "")
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return CreateSnapResponse{}, err
	}
	defer resp.Body.Close()
	var out struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
		StatusMsg   string `json:"status_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CreateSnapResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CreateSnapResponse{}, fmt.Errorf("midtrans create snap failed: %s", out.StatusMsg)
	}
	return CreateSnapResponse{Token: out.Token, RedirectURL: out.RedirectURL}, nil
}

func VerifyMidtransSignature(orderID, statusCode, grossAmount, signatureKey, serverKey string) bool {
	raw := orderID + statusCode + grossAmount + serverKey
	sum := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(sum[:])
	return strings.EqualFold(strings.TrimSpace(signatureKey), strings.TrimSpace(expected))
}

