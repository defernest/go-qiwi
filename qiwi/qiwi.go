package qiwi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL = "https://api.qiwi.com/"
)

type Client struct {
	client *http.Client

	BaseURL   *url.URL
	UserAgent string
	APIKey    string

	Invoice *InvoiceService
	Status  *StatusInvoiceService
	Cancel  *CancelInvoiceService
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{client: httpClient}
	c.BaseURL, _ = url.Parse(baseURL)
	c.Invoice = &InvoiceService{client: c}
	c.Status = &StatusInvoiceService{client: c}
	c.Cancel = &CancelInvoiceService{client: c}
	return c
}

type ErrorResponse struct {
	ServiceName string    `json:"serviceName"`
	ErrorCode   string    `json:"errorCode"`
	Description string    `json:"description"`
	UserMessage string    `json:"userMessage"`
	DateTime    time.Time `json:"dateTime"`
	TraceID     string    `json:"traceId"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("qiwi API error: \nService Name: %s\nError Code: %s\nDescription: %s\nUser Message: %s\nDateTime: %s\nTraceID: %s\n",
		e.ServiceName, e.ErrorCode, e.Description, e.UserMessage, e.DateTime, e.TraceID)
}

type SuccessResponse struct {
	SiteID string `json:"siteId"`
	BillID string `json:"billId"`
	Amount struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"amount"`
	Status struct {
		Value           string    `json:"value"`
		ChangedDateTime time.Time `json:"changedDateTime"`
	} `json:"status"`
	Customer struct {
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
	Comment            string    `json:"comment"`
	CreationDateTime   time.Time `json:"creationDateTime"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	PayURL             string    `json:"payUrl"`
}

func (c *Client) makeRequest(method, path string, body io.Reader) (*http.Request, error) {
	rel := &url.URL{Path: path}
	url := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}
	if body != nil || method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	return req, nil
}
func (c *Client) do(req *http.Request, body interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.Body != http.NoBody {
			APIError := new(ErrorResponse)
			err = json.NewDecoder(resp.Body).Decode(APIError)
			if err != nil {
				return nil, err
			}
			return nil, APIError
		}
		return nil, fmt.Errorf("qiwi API http error: %s", resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(body)
	if err != nil {
		return nil, fmt.Errorf("error in json decode %w", err)
	}
	return resp, nil
}
