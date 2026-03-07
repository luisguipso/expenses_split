package service

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/lguilherme/contas/internal/domain"
)

type smtpEmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailService(host string, port int, username, password, from string) domain.EmailService {
	return &smtpEmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpEmailService) SendVerificationCode(to, code string) error {
	subject := "Contas - Código de verificação"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #1e40af;">Contas</h2>
  <p>Olá! Seu código de verificação é:</p>
  <div style="background: #f1f5f9; border-radius: 8px; padding: 20px; text-align: center; margin: 20px 0;">
    <span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1e40af;">%s</span>
  </div>
  <p style="color: #64748b; font-size: 14px;">Este código expira em 15 minutos. Se você não solicitou este código, ignore este email.</p>
</body>
</html>`, code)

	return s.send(to, subject, body)
}

func (s *smtpEmailService) SendPasswordResetLink(to, resetLink string) error {
	subject := "Contas - Recuperação de senha"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #1e40af;">Contas</h2>
  <p>Olá! Recebemos uma solicitação para redefinir sua senha.</p>
  <p>Clique no botão abaixo para criar uma nova senha:</p>
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s" style="background: #1e40af; color: #ffffff; padding: 14px 28px; border-radius: 8px; text-decoration: none; font-weight: bold; font-size: 16px;">Redefinir senha</a>
  </div>
  <p style="color: #64748b; font-size: 14px;">Este link expira em 30 minutos. Se você não solicitou a redefinição de senha, ignore este email.</p>
  <p style="color: #94a3b8; font-size: 12px;">Se o botão não funcionar, copie e cole este link no seu navegador: %s</p>
</body>
</html>`, resetLink, resetLink)

	return s.send(to, subject, body)
}

func (s *smtpEmailService) send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	slog.Info("sending email", "to", to, "smtp_host", addr)

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		slog.Error("smtp send failed", "error", err, "to", to, "smtp_host", addr)
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
