package qiwi

import (
	"errors"
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

func TestGetSuccess(t *testing.T) {
	assert := assert.New(t)
	amount := 1
	currency := "RUB"
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	expTime := time.Date(2021, 07, 25, 11, 11, 01, 0, time.UTC)
	json := []byte(fmt.Sprintf(`{"siteId":"0dwgg9-00","billId":"%s","amount":{"currency":"%s","value":"%d"},"status":{"value":"REJECTED","changedDateTime":"2021-07-24T11:57:46.541+03:00"},"creationDateTime":"2021-07-24T11:17:37.717+03:00","expirationDateTime":"%s","payUrl":"https://oplata.qiwi.com/form/?invoice_uid=66057202-07ee-4b49-8ec8-2a344cb90226"}`, billID, currency, amount, expTime.Format(time.RFC3339)))
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
	client.Status = &StatusInvoiceService{client: &client}

	resp, err := client.Status.Get(billID)
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
	assert.Equal(expTime, resp.ExpirationDateTime)
	assert.Contains(resp.PayURL, "qiwi.com")
}

func TestGetWrongStatus(t *testing.T) {
	assert := assert.New(t)
	amount := 1
	currency := "RUB"
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	testCases := []struct {
		status string
		err    error
	}{
		{"WAITING", nil},
		{"PAID", nil},
		{"REJECTED", nil},
		{"EXPIRED", nil},
		{"paid", nil},
		{"wrongstatus", errors.New("api return wrong status paid")},
	}
	for _, test := range testCases {
		expTime := time.Date(2021, 07, 25, 11, 11, 01, 0, time.UTC)
		json := []byte(fmt.Sprintf(`{"siteId":"0dwgg9-00","billId":"%s","amount":{"currency":"%s","value":"%d"},"status":{"value":"%s","changedDateTime":"2021-07-24T11:57:46.541+03:00"},"creationDateTime":"2021-07-24T11:17:37.717+03:00","expirationDateTime":"%s","payUrl":"https://oplata.qiwi.com/form/?invoice_uid=66057202-07ee-4b49-8ec8-2a344cb90226"}`, billID, currency, amount, test.status, expTime.Format(time.RFC3339)))
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
		client.Status = &StatusInvoiceService{client: &client}

		resp, err := client.Status.Get(billID)
		if err != nil {
			log.Fatalln(err)
			assert.EqualError(err, test.err.Error())
		}
		assert.Equal(test.status, resp.Status.Value)
	}
}
