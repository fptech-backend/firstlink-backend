package mailer

import (
	"certification/config"
	"certification/logger"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/go-gomail/gomail"
)

func SetUpSMTP(from string) (*gomail.Dialer, error) {

	dialer := gomail.NewDialer(config.SMTP_HOST, config.SMTP_PORT, config.SMTP_FROM, config.SMTP_PASSWORD)

	// Set up TLS configuration
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         config.SMTP_HOST,
	}

	return dialer, nil
}

func SendEmail(bodyhtml, subject string, to_emails []string, file *multipart.FileHeader) error {
	dialer, err := SetUpSMTP(config.SMTP_FROM)
	if err != nil {
		return err
	}

	// Set up email message
	for _, to := range to_emails {
		m := gomail.NewMessage()
		m.SetHeader("From", config.SMTP_FROM)
		m.SetHeader("To", to)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", bodyhtml)

		// Attach the file if it exists
		if file != nil {
			attachment, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer attachment.Close()

			m.Attach(file.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := io.Copy(w, attachment)
				return err
			}))
		}

		// Send email individually
		if err := dialer.DialAndSend(m); err != nil {
			logger.Log.Errorf("failed to send email to %v. %v", to, err)
			continue
		}
	}
	return nil
}
