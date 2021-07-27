package qiwi

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type LifeTime struct {
	lt time.Time
}

func (lt LifeTime) LifeTime() (ISOTime time.Time, err error) {
	return lt.lt.UTC().Truncate(time.Microsecond), nil
}

func TestNewInvoiceRequest(t *testing.T) {
	assert := assert.New(t)

	timenow := time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second)

	var tests = []struct {
		amount   int
		currency string
		expTime  LifeTime
		result   *InvoiceRequest
		err      string
	}{
		{
			amount:   10,
			currency: "RUB",
			expTime:  LifeTime{lt: timenow},
			result: &InvoiceRequest{Amount: struct {
				Currency string "json:\"currency\""
				Value    string "json:\"value\""
			}{Currency: "RUB", Value: "10"},
				ExpirationDateTime: timenow},
			err: "",
		},
		{
			amount:   10,
			currency: "KZT",
			expTime:  LifeTime{lt: timenow},
			result: &InvoiceRequest{Amount: struct {
				Currency string "json:\"currency\""
				Value    string "json:\"value\""
			}{Currency: "KZT", Value: "10"},
				ExpirationDateTime: timenow},
			err: "",
		},
		{
			amount:   -10,
			currency: "RUB",
			expTime:  LifeTime{lt: time.Date(2021, 11, 1, 1, 0, 0, 0, time.UTC)},
			result:   &InvoiceRequest{},
			err:      "amount: cannot be negative or null (-10)",
		},
		{
			amount:   10,
			currency: "RU",
			expTime:  LifeTime{lt: time.Date(2021, 11, 1, 1, 0, 0, 0, time.UTC)},
			result:   &InvoiceRequest{},
			err:      "incorrect currency value (RU)",
		},
		{
			amount:   10,
			currency: "RUB",
			expTime:  LifeTime{lt: time.Date(2020, 11, 1, 1, 0, 0, 0, time.UTC)},
			result:   &InvoiceRequest{},
			err:      fmt.Sprintf("ExpirationDateTime [expTime] cannot be in the past (%s)", time.Date(2020, 11, 1, 1, 0, 0, 0, time.UTC)),
		},
	}
	for _, test := range tests {
		ir, err := NewInvoiceRequest(test.amount, test.currency, test.expTime)
		if err != nil {
			assert.EqualError(err, test.err, test)
		} else {
			assert.Equal(test.result, ir,
				fmt.Sprintf("exp: %s actual: %s",
					test.result.ExpirationDateTime, ir.ExpirationDateTime))
		}
	}
}

func TestIssueSuccess(t *testing.T) {
	assert := assert.New(t)
	amount := 1
	currency := "RUB"
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	expTime := LifeTime{lt: time.Now().Add(1 * time.Hour).UTC()}
	json := []byte(fmt.Sprintf(`{"siteId":"0dwgg9-00","billId":"%s","amount":{"currency":"%s","value":"%d"},"status":{"value":"REJECTED","changedDateTime":"2021-07-24T11:57:46.541+03:00"},"creationDateTime":"2021-07-24T11:17:37.717+03:00","expirationDateTime":"%s","payUrl":"https://oplata.qiwi.com/form/?invoice_uid=66057202-07ee-4b49-8ec8-2a344cb90226"}`, billID, currency, amount, expTime.lt.Format(time.RFC3339)))
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(json)
	}))
	url, err := url.Parse(server.URL)
	if err != nil {
		log.Fatalln(err)
	}
	url.Path = "/partner/bill/v1/bills/" + billID
	client := Client{
		client:  server.Client(),
		BaseURL: url,
	}
	client.Invoice = &InvoiceService{client: &client}
	irr, err := NewInvoiceRequest(1, currency, expTime)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := client.Invoice.Issue(irr)
	if err != nil {
		log.Fatalln(err)
	}
	respamount, err := strconv.Atoi(resp.Amount.Value)
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(billID, resp.BillID)
	assert.Equal(currency, resp.Amount.Currency)
	assert.Equal(amount, respamount)
	assert.Equal(expTime.lt.Truncate(time.Second), resp.ExpirationDateTime, fmt.Sprintf("exp: %s actual: %s", expTime.lt, resp.ExpirationDateTime))
	assert.Contains(resp.PayURL, "qiwi.com")
}

func TestIssueAPIError(t *testing.T) {
	assert := assert.New(t)
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	json := []byte(`
		{
			"serviceName": "invoicing-api",
			"errorCode": "api.invoice.not.found",
			"description": "Invoice not found",
			"userMessage": "Invoice not found",
			"dateTime": "2021-01-18T14:39:54.265+03:00",
			"traceId": "bc6bb6e7c5cf5beb"
		}
	`)
	errDT, err := time.Parse("2006-01-02T15:04:05.999999-07:00", "2021-01-18T14:39:54.265+03:00")
	if err != nil {
		log.Fatalln(err)
	}
	qiwiErr := &ErrorResponse{
		ServiceName: "invoicing-api",
		ErrorCode:   "api.invoice.not.found",
		UserMessage: "Invoice not found",
		Description: "Invoice not found",
		DateTime:    errDT,
		TraceID:     "bc6bb6e7c5cf5beb",
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(json)
	}))
	url, err := url.Parse(server.URL)
	if err != nil {
		log.Fatalln(err)
	}
	url.Path = "/partner/bill/v1/bills/" + billID
	client := Client{
		client:  server.Client(),
		BaseURL: url,
	}
	client.Invoice = &InvoiceService{client: &client}
	irr, err := NewInvoiceRequest(1, "RUB", LifeTime{lt: time.Now()})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = client.Invoice.Issue(irr)
	assert.EqualError(err, qiwiErr.Error())
}
