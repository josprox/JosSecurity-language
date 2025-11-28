package core

import (
	"fmt"
	"net/smtp"
)

// SmtpClient Implementation
func (r *Runtime) executeSmtpClientMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "auth":
		if len(args) == 2 {
			instance.Fields["user"] = args[0]
			instance.Fields["pass"] = args[1]
		}
		return instance
	case "secure":
		if len(args) == 1 {
			instance.Fields["secure"] = args[0]
		}
		return instance
	case "send":
		if len(args) >= 3 {
			to := args[0].(string)
			subject := args[1].(string)
			body := args[2].(string)

			// Defaults
			host := "smtp.gmail.com"
			port := "587"
			if h, ok := r.Env["MAIL_HOST"]; ok {
				host = h
			}
			if p, ok := r.Env["MAIL_PORT"]; ok {
				port = p
			}

			user := ""
			pass := ""
			if u, ok := instance.Fields["user"]; ok {
				user = u.(string)
			}
			if p, ok := instance.Fields["pass"]; ok {
				pass = p.(string)
			}

			msg := []byte("From: " + user + "\r\n" +
				"To: " + to + "\r\n" +
				"Subject: " + subject + "\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
				"\r\n" +
				body + "\r\n")

			auth := smtp.PlainAuth("", user, pass, host)
			err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
			if err != nil {
				fmt.Printf("[SmtpClient] Error enviando correo: %v\n", err)
				return false
			}
			fmt.Println("[SmtpClient] Correo enviado exitosamente a " + to)
			return true
		}
	}
	return nil
}
