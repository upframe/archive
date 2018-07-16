package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/upframe/api"
)

// Config ...
type Config struct {
	Host       string
	Port       string
	ServerName string
	Auth       smtp.Auth
	TLSConfig  *tls.Config
}

// EmailService ...
type EmailService struct {
	TemplatesPath    string
	FromDefaultEmail string
	SMTP             *Config
}

// InitSMTP configures the email variables
func InitSMTP(user, pass, host, port string) *EmailService {
	s := &EmailService{
		SMTP: &Config{
			Host:       host,
			Port:       port,
			ServerName: host + ":" + port,
			Auth:       smtp.PlainAuth("", user, pass, host),
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         host,
			},
		},
	}

	return s
}

var (
	funcs = template.FuncMap{
		"CSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"HTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"URL": func(s string) template.URL {
			return template.URL(s)
		},
	}
)

// UseTemplate adds the template to the email and renders it to the Body field
// If 'name' doesn't correspond to any file in the templates folder, it must be
// a template body.
func (s *EmailService) UseTemplate(e *api.Email, data interface{}, name string) error {
	// Opens the template file and checks if there is any error
	page, err := ioutil.ReadFile(filepath.Clean(s.TemplatesPath + name + ".html"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var raw string
	if os.IsNotExist(err) {
		raw = name
	} else {
		raw = string(page)
	}

	// Create the template with the content of the file
	tpl, err := template.New("template").Funcs(funcs).Parse(raw)

	if err != nil {
		return err
	}

	// Creates a buffer to render the template into it
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, data)

	if err != nil {
		return err
	}

	e.Body = buf.String()
	return nil
}

// Send sends the email
func (s *EmailService) Send(e *api.Email) error {
	if e.From.Address == "" {
		e.From.Address = s.FromDefaultEmail
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = e.From.String()
	headers["To"] = e.To.String()
	headers["Subject"] = e.Subject
	headers["Content-Type"] = "text/html; charset=utf-8"

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + e.Body

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", s.SMTP.ServerName, s.SMTP.TLSConfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, s.SMTP.Host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(s.SMTP.Auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(e.From.Address); err != nil {
		return err
	}

	if err = c.Rcpt(e.To.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	return nil
}
