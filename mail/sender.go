package mail

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"net/mail"
	"net/smtp"
)

type Client struct {
	Username, Password string
	LocalName          string
}

func (c Client) Send(host string, port uint16, tlsConfig *tls.Config, from mail.Address, to mail.Address, subject string, body string) error {
	if len(c.Username) == 0 {
		return errors.New("Username has to be set")
	}
	if len(c.Password) == 0 {
		return errors.New("Password has to be set")
	}

	serverName := fmt.Sprintf("%s:%d", host, port)
	headers := setupHeaders(from, to, subject)
	message := setupMessage(headers, body)
	auth := smtp.PlainAuth("", c.Username, c.Password, host)

	client, err := smtp.Dial(serverName)
	if err != nil {
		return wrapError(err)
	}
	defer client.Close()

	if len(c.LocalName) != 0 {
		if err = client.Hello(c.LocalName); err != nil {
			return wrapError(err)
		}
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return wrapError(err)
	}
	_, ok := client.TLSConnectionState()
	if !ok {
		err := errors.New("TLS is not ok")
		return wrapError(err)
	}
	// Auth
	if err = client.Auth(auth); err != nil {
		return wrapError(err)
	}
	// From
	if err = client.Mail(from.Address); err != nil {
		return wrapError(err)
	}
	// To
	if err = client.Rcpt(to.Address); err != nil {
		return wrapError(err)
	}
	// Data
	w, err := client.Data()
	if err != nil {
		return wrapError(err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return wrapError(err)
	}
	err = w.Close()
	if err != nil {
		return wrapError(err)
	}
	err = client.Quit()
	if err != nil {
		return wrapError(err)
	}

	return nil
}

func wrapError(err error) error {
	return errors.Wrap(err, "Client.Send")
}

func setupMessage(headers map[string]string, body string) string {
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	return message
}

func setupHeaders(from mail.Address, to mail.Address, subject string) map[string]string {
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject
	return headers
}
