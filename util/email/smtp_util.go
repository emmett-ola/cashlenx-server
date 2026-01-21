package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"time"

	"github.com/macar-x/cashlenx-server/util"
)

// Email represents an email message
type Email struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail sends an email using the configured SMTP server with retry logic
func SendEmail(email Email) error {
	// Get configuration
	host := util.GetConfigByKey("smtp.host")
	portStr := util.GetConfigByKey("smtp.port")
	username := util.GetConfigByKey("smtp.username")
	password := util.GetConfigByKey("smtp.password")
	fromAddress := util.GetConfigByKey("smtp.from_address")
	fromName := util.GetConfigByKey("smtp.from_name")
	
	maxRetriesStr := util.GetConfigByKey("smtp.max_retries")
	retryIntervalStr := util.GetConfigByKey("smtp.retry_interval")

	// Validate configuration
	if host == "" || portStr == "" || username == "" || password == "" || fromAddress == "" {
		util.Logger.Errorw("SMTP configuration is missing", "host", host, "port", portStr, "username", username)
		return fmt.Errorf("SMTP configuration is missing")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		util.Logger.Errorw("Invalid SMTP port", "port", portStr, "error", err)
		return err
	}

	maxRetries := 3
	if maxRetriesStr != "" {
		if val, err := strconv.Atoi(maxRetriesStr); err == nil {
			maxRetries = val
		}
	}

	retryInterval := 1000 // ms
	if retryIntervalStr != "" {
		if val, err := strconv.Atoi(retryIntervalStr); err == nil {
			retryInterval = val
		}
	}

	// Prepare email content
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	// Construct headers
	contentType := "text/plain"
	if email.IsHTML {
		contentType = "text/html"
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	headers["To"] = email.To[0] // Simplify for single recipient
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=\"UTF-8\"", contentType)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + email.Body

	// Retry loop
	var sendErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			util.Logger.Infow("Retrying email send", "attempt", i+1, "to", email.To)
			time.Sleep(time.Duration(retryInterval) * time.Millisecond)
		}

		// Try to send email
		// Note: smtp.SendMail uses the default TLS configuration if StartTLS is supported
		// For explicit TLS control or STARTTLS issues, might need custom implementation
		// This is a basic implementation compatible with most providers like Mailgun
		sendErr = sendMailWithTLS(addr, auth, fromAddress, email.To, []byte(message), host)
		
		if sendErr == nil {
			util.Logger.Infow("Email sent successfully", "to", email.To, "subject", email.Subject)
			return nil
		}

		util.Logger.Errorw("Failed to send email", "attempt", i+1, "error", sendErr)
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, sendErr)
}

// sendMailWithTLS sends an email with proper TLS handling
func sendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	// Custom implementation to ensure TLS/StartTLS works correctly
	// This avoids common issues with some providers
	
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: host}
		if err = client.StartTLS(config); err != nil {
			return err
		}
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
