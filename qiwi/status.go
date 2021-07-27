package qiwi

import (
	"fmt"
)

type StatusInvoiceService struct {
	client *Client
}

func (s *StatusInvoiceService) Get(billID string) (*SuccessResponse, error) {
	req, err := s.client.makeRequest("GET", fmt.Sprintf("/partner/bill/v1/bills/%s", billID), nil)
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
