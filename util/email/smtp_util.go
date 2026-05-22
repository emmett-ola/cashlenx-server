package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"sync"
	"time"

	"github.com/macar-x/cashlenx-server/util"
)

// SMTPService defines the interface for sending emails
type SMTPService interface {
	SendEmail(to []string, subject, body string, isHTML bool) error
	SendPurposeEmail(to []string, subject, body string, isHTML bool, purpose string, ipAddress string) error
	IsConfigured() bool
}

// smtpServiceImpl implements SMTPService
type smtpServiceImpl struct {
	host          string
	port          int
	username      string
	password      string
	fromAddress   string
	fromName      string
	maxRetries    int
	retryInterval int
	configured    bool
	mu            sync.RWMutex
}

var (
	instance *smtpServiceImpl
	once     sync.Once
)

// GetService returns the singleton SMTP service instance
func GetService() SMTPService {
	once.Do(func() {
		instance = &smtpServiceImpl{}
		instance.loadConfig()
	})
	return instance
}

// ReloadConfig reloads the SMTP configuration (useful if config changes at runtime)
func ReloadConfig() {
	if instance != nil {
		instance.loadConfig()
	}
}

func (s *smtpServiceImpl) loadConfig() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.host = util.GetConfigByKey("smtp.host")
	portStr := util.GetConfigByKey("smtp.port")
	s.username = util.GetConfigByKey("smtp.username")
	s.password = util.GetConfigByKey("smtp.password")
	s.fromAddress = util.GetConfigByKey("smtp.from_address")
	s.fromName = util.GetConfigByKey("smtp.from_name")

	maxRetriesStr := util.GetConfigByKey("smtp.max_retries")
	retryIntervalStr := util.GetConfigByKey("smtp.retry_interval")

	// Validate minimal required config
	if s.host == "" || portStr == "" || s.username == "" || s.password == "" || s.fromAddress == "" {
		s.configured = false
		util.Logger.Warn("SMTP configuration is incomplete. Email features will be disabled.")
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		s.configured = false
		util.Logger.Errorw("Invalid SMTP port configuration", "port", portStr, "error", err)
		return
	}
	s.port = port

	// Optional config with defaults
	s.maxRetries = 3
	if maxRetriesStr != "" {
		if val, err := strconv.Atoi(maxRetriesStr); err == nil {
			s.maxRetries = val
		}
	}

	s.retryInterval = 1000 // ms
	if retryIntervalStr != "" {
		if val, err := strconv.Atoi(retryIntervalStr); err == nil {
			s.retryInterval = val
		}
	}

	s.configured = true
	util.Logger.Info("SMTP service configured successfully")
}

func (s *smtpServiceImpl) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configured
}

// SendEmail sends an email using the configured SMTP server with retry logic
func (s *smtpServiceImpl) SendEmail(to []string, subject, body string, isHTML bool) error {
	if !s.IsConfigured() {
		return fmt.Errorf("SMTP service is not configured")
	}

	s.mu.RLock()
	// Copy config to local vars to avoid holding lock during network I/O
	host := s.host
	port := s.port
	username := s.username
	password := s.password
	fromAddress := s.fromAddress
	fromName := s.fromName
	maxRetries := s.maxRetries
	retryInterval := s.retryInterval
	s.mu.RUnlock()

	// Prepare email content
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	// Construct headers
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	if len(to) > 0 {
		headers["To"] = to[0] // Simplify for single recipient
	}
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=\"UTF-8\"", contentType)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Retry loop
	var sendErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			util.Logger.Infow("Retrying email send", "attempt", i+1, "to", to)
			time.Sleep(time.Duration(retryInterval) * time.Millisecond)
		}

		sendErr = sendMailWithTLS(addr, auth, fromAddress, to, []byte(message), host)

		if sendErr == nil {
			util.Logger.Infow("Email sent successfully", "to", to, "subject", subject)
			return nil
		}

		util.Logger.Errorw("Failed to send email", "attempt", i+1, "error", sendErr)
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, sendErr)
}

// SendPurposeEmail sends a purpose-scoped email after applying SMTP abuse limits.
func (s *smtpServiceImpl) SendPurposeEmail(to []string, subject, body string, isHTML bool, purpose string, ipAddress string) error {
	if err := CheckAndRecordPurposeEmailAllowance(purpose, ipAddress, to); err != nil {
		return err
	}
	return s.SendEmail(to, subject, body, isHTML)
}

// Helper function: SendEmail (Backward compatibility wrapper)
func SendEmail(emailStruct struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}) error {
	return GetService().SendEmail(emailStruct.To, emailStruct.Subject, emailStruct.Body, emailStruct.IsHTML)
}

// sendMailWithTLS sends an email with proper TLS handling
func sendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
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
