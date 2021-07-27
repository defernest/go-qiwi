package qiwi

import (
	"fmt"
)

type CancelInvoiceService struct {
	client *Client
}

func (s *CancelInvoiceService) Cancel(billID string) (*SuccessResponse, error) {
	req, err := s.client.makeRequest("POST", fmt.Sprintf("/partner/bill/v1/bills/%s/reject", billID), nil)
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
