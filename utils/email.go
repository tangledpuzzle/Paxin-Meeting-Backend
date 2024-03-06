package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/k3a/html2text"

	"gopkg.in/gomail.v2"
)

type EmailData struct {
	URL       string
	FirstName string
	Subject   string
}

type ReqCat struct {
	Name    string
	Descr   string
	Subject string
}

type ReqCity struct {
	Name    string
	Descr   string
	Subject string
}

type ComplainUser struct {
	Type    string
	Name    string
	Descr   string
	Subject string
}

type ComplainPost struct {
	Type    string
	Name    string
	Descr   string
	Subject string
}

// ? Email template parser

func ParseTemplateDir(dir string) (*template.Template, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return template.ParseFiles(paths...)
}

func SendEmail(user *models.User, data interface{}, emailTemplatePrefix string, language string) {
	config, err := initializers.LoadConfig(".")

	if err != nil {
		log.Fatal("could not load config", err)
	}

	// Sender data.
	from := config.EmailFrom
	smtpPass := config.SMTPPass
	smtpUser := config.SMTPUser
	to := user.Email
	smtpHost := config.SMTPHost
	smtpPort := config.SMTPPort

	var body bytes.Buffer

	var emailTemplate string
	switch data.(type) {
	case *EmailData:
		emailTemplate = emailTemplatePrefix + "_" + language + ".html"
	case *ReqCat:
		emailTemplate = emailTemplatePrefix + "_" + language + ".html"
	case *ReqCity:
		emailTemplate = emailTemplatePrefix + "_" + language + ".html"
	case *ComplainUser:
		emailTemplate = emailTemplatePrefix + "_" + language + ".html"
	case *ComplainPost:
		emailTemplate = emailTemplatePrefix + "_" + language + ".html"
	default:
		log.Fatal("Unsupported email data type")
	}

	template, err := ParseTemplateDir("templates")
	if err != nil {
		log.Fatal("Could not parse template", err)
	}

	template.ExecuteTemplate(&body, emailTemplate, data)

	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	switch data := data.(type) {
	case *EmailData:
		m.SetHeader("Subject", data.Subject)
	case *ReqCat:
		m.SetHeader("Subject", data.Subject)
	case *ReqCity:
		m.SetHeader("Subject", data.Subject)
	case *ComplainUser:
		m.SetHeader("Subject", data.Subject)
	case *ComplainPost:
		m.SetHeader("Subject", data.Subject)
	default:
		log.Println("Unsupported email data type")
	}
	m.SetBody("text/html", body.String())
	m.AddAlternative("text/plain", html2text.HTML2Text(body.String()))

	fmt.Println(data)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send Email
	if err := d.DialAndSend(m); err != nil {
		log.Fatal("Could not send email: ", err)
	}

}
