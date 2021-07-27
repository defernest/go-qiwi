package qiwi

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIssueHTTPError(t *testing.T) {
	assert := assert.New(t)
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "application/json")
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
	assert.EqualError(err, "qiwi API http error: 500 Internal Server Error")
}

func TestGetHTTPError(t *testing.T) {
	assert := assert.New(t)
	billID := "cc961e8d-d4d6-4f02-b737-2297e51fb48e"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "application/json")
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

	_, err = client.Status.Get(billID)
	assert.EqualError(err, "qiwi API http error: 500 Internal Server Error")
}
