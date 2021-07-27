package qiwi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type InvoiceRequest struct {
	Amount struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"amount"`
	Comment            string    `json:"comment"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	Customer           struct {
		Phone   string `json:"phone"`
		Email   string `json:"email"`
		Account string `json:"account"`
	} `json:"customer"`
	CustomFields struct {
		PaySourcesFilter string `json:"paySourcesFilter"`
		ThemeCode        string `json:"themeCode"`
		YourParam1       string `json:"yourParam1"`
		YourParam2       string `json:"yourParam2"`
	} `json:"customFields"`
}

// Lifetime - ISO 8601 UTC±0:00
type Lifetime interface {
	LifeTime() (ISOTime time.Time, err error)
}
type ExpTime struct {
	Hours int
}

//2018-02-05T17:16:58.033Z
func (et ExpTime) LifeTime() (ISOTime time.Time, err error) {
	if et.Hours == 0 {
		et.Hours = 1
	}
	if et.Hours >= 1 {
		ISOTime = time.Now().
			Add(time.Duration(et.Hours) * time.Hour).
			UTC().Truncate(time.Millisecond)
		return
	}
	return time.Time{}, fmt.Errorf("%#v cannot be less than one hour", et)
}

func NewInvoiceRequest(amount int, currency string, expTime Lifetime) (*InvoiceRequest, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount: cannot be negative or null (%d)", amount)
	}
	a := strconv.Itoa(amount)
	currency = strings.TrimSpace(currency)
	if strings.Compare(currency, "RUB") != 0 && strings.Compare(currency, "KZT") != 0 {
		return nil, fmt.Errorf("incorrect currency value (%s)", currency)
	}
	lifetime, err := expTime.LifeTime()
	if err != nil {
		return nil, err
	}
	if lifetime.Unix() < time.Now().UTC().Unix() {
		return nil, fmt.Errorf("ExpirationDateTime [expTime] cannot be in the past (%s)", lifetime)
	}
	return &InvoiceRequest{
		Amount: struct {
			Currency string "json:\"currency\""
			Value    string "json:\"value\""
		}{currency, a},
		ExpirationDateTime: lifetime,
	}, nil
}

type InvoiceService struct {
	client *Client
}

func (s InvoiceService) Issue(ir *InvoiceRequest) (*SuccessResponse, error) {
	uuid := uuid.New()
	b, err := json.Marshal(ir)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(b)
	req, err := s.client.makeRequest("PUT", fmt.Sprintf("/partner/bill/v1/bills/%s", uuid), body)
	if err != nil {
		return nil, err
	}
	responseBody := new(SuccessResponse)
	_, err = s.client.do(req, responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
}
