package email

import (
	mailgungo "github.com/mailgun/mailgun-go"
)

type Client struct {
	Client *mailgungo.MailgunImpl
}

func NewClient(apiKey, domain string) *Client {
	client := mailgungo.NewMailgun(domain, apiKey)
	return &Client{
		Client: client,
	}
}
