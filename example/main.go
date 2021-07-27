package main

import (
	"fmt"
	"log"
	"os"

	"github.com/defernest/go-qiwi/qiwi"
)

func main() {
	c := qiwi.NewClient(nil)
	c.APIKey = os.Getenv("QIWI_KEY")
	// Make invoice request with exp time 5 hours
	ir, err := qiwi.NewInvoiceRequest(1, "RUB", qiwi.ExpTime{Hours: 3})
	if err != nil {
		log.Fatalln(err)
	}
	//Request payment
	resp, err := c.Invoice.Issue(ir)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Bill UUID: %s | Pay URL: %s", resp.BillID, resp.PayURL)
	// Get payment status
	resp, err = c.Status.Get(resp.BillID)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Payment [%s] status: %s", resp.BillID, resp.Status.Value)
	// Cansel unpayed payment
	resp, err = c.Cancel.Cancel(resp.BillID)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Payment [%s] %s", resp.BillID, resp.Status.Value)
}
