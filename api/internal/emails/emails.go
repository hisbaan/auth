package emails

import (
	_ "embed"
	"html/template"
	"log"
	"strings"

	"github.com/resend/resend-go/v3"
)

//go:embed templates/forgot-password.html
var forgotPasswordTemplate string

type EmailService struct {
	client      *resend.Client
	from        string
	frontendURL string
	serviceName string
}

func NewEmailService(resendAPIKey string, from string, frontendURL string, serviceName string) (*EmailService, error) {
	return &EmailService{
		client:      resend.NewClient(resendAPIKey),
		from:        from,
		frontendURL: frontendURL,
		serviceName: serviceName,
	}, nil
}

func (s *EmailService) SendEmail(to []string, html string, subject string) {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      to,
		Html:    html,
		Subject: subject,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("[ERROR] Issue sending email %+v\n", params)
		return
	}
	// TODO log email sending, then add a webhook for email status reporting
}

func (s *EmailService) SendForgotPasswordEmail(to string, username string, resetToken string) {
	resetURL := s.frontendURL + "/reset-password?token=" + resetToken

	tmpl, err := template.New("forgot-password").Parse(forgotPasswordTemplate)
	if err != nil {
		log.Printf("[ERROR] Failed to parse forgot password template: %v", err)
		return
	}

	type forgotPasswordData struct {
		Username    string
		ResetLink   string
		AuthURL     string
		ServiceName string
	}

	data := forgotPasswordData{
		Username:    username,
		ResetLink:   resetURL,
		AuthURL:     s.frontendURL,
		ServiceName: s.serviceName,
	}

	var htmlBuilder strings.Builder
	if err := tmpl.Execute(&htmlBuilder, data); err != nil {
		log.Printf("[ERROR] Failed to execute forgot password template: %v", err)
		return
	}

	s.SendEmail([]string{to}, htmlBuilder.String(), "Reset your password - "+s.serviceName)
}
