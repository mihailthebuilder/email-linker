package main

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Sendgrid struct {
	client *sendgrid.Client
	from   *mail.Email
}

func newSendGridClient() *Sendgrid {
	client := sendgrid.NewSendClient(getEnv("SENDGRID_API_KEY"))
	from := mail.NewEmail("Sender", getEnv("EMAILS_FROM"))
	return &Sendgrid{client: client, from: from}
}

type SendEmailRequest struct {
	Email   string
	Subject string
	Content string
}

func (s *Sendgrid) SendEmail(request SendEmailRequest) error {
	to := mail.NewEmail("Recipient", request.Email)
	message := mail.NewSingleEmailPlainText(s.from, request.Subject, to, request.Content)
	response, err := s.client.Send(message)
	if err != nil {
		return fmt.Errorf("error sending email %#v: %s", request, err)
	}

	if response.StatusCode != 202 {
		return fmt.Errorf("status code %d when sending email %#v: %s", response.StatusCode, request, err)
	}

	return nil
}
