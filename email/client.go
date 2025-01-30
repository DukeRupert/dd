package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const postmarkAPIURL = "https://api.postmarkapp.com/email"

type Client struct {
	serverToken string
	fromEmail   string
	httpClient  *http.Client
	enabled     bool
}

type PostmarkRequest struct {
	From          string `json:"From"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	TextBody      string `json:"TextBody"`
	HtmlBody      string `json:"HtmlBody"`
	MessageStream string `json:"MessageStream"`
}

func NewClient(serverToken, fromEmail string) (*Client, error) {
	if serverToken == "" {
		return nil, fmt.Errorf("missing required POSTMARK_SERVER_TOKEN")
	}
	if fromEmail == "" {
		return nil, fmt.Errorf("missing required FROM_EMAIL")
	}

	return &Client{
		serverToken: serverToken,
		fromEmail:   fromEmail,
		httpClient:  &http.Client{},
		enabled:     true,
	}, nil
}

func (c *Client) SendEmail(to, subject, textBody, htmlBody string) error {
	if !c.enabled {
		// Log the email content instead of sending
		fmt.Printf("Email would have been sent:\nTo: %s\nSubject: %s\nBody: %s\n", 
			to, subject, textBody)
		return nil
	}

	req := PostmarkRequest{
		From:          c.fromEmail,
		To:            to,
		Subject:       subject,
		TextBody:      textBody,
		HtmlBody:      htmlBody,
		MessageStream: "outbound",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", postmarkAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Postmark-Server-Token", c.serverToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email service returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}