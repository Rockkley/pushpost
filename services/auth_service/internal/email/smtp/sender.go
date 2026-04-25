package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	AppName  string
}

type Sender struct {
	cfg  Config
	auth smtp.Auth
	addr string
}

func NewSender(cfg Config) *Sender {
	return &Sender{
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
		addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
}

func (s *Sender) SendOTP(ctx context.Context, to, code string) error {
	msg := s.buildMessage(to, code)
	if err := smtp.SendMail(s.addr, s.auth, s.cfg.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send otp to %s: %w", to, err)
	}

	return nil
}

func (s *Sender) buildMessage(to, code string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", s.cfg.From))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: Confirm your %s account\r\n", s.cfg.AppName))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf(
		"Your confirmation code: %s\n\nThe code is valid for 5 minutes.\n\nIf you did not register, ignore this email.",
		code,
	))

	return b.String()
}
