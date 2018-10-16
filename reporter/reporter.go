package reporter

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// convert to env or flags
const (
	apiHost = "api.sendgrid.com"
	name    = "jrieck1991"
)

// initialize variables for sending email
func initVars() (from, to, key string) {
	from = os.Getenv("EMAIL_FROM")
	to = os.Getenv("EMAIL_TO")
	key = os.Getenv("EMAIL_API_KEY_V3")

	if containsEmpty(from, to, key) {
		log.Fatalln(errors.New("error: unintialized email variables, exiting"))
	}

	return from, to, key
}

// check if email variables are empty
func containsEmpty(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return true
		}
	}
	return false
}

// Email body template
const body = `
Attached to this email are results from crawling these search engines:
{{range $key, $value := .}}
	<p>{{$key}}: {{$value}}</p>
{{end}}
`

// createHTMLBody creates template for email body
func createHTMLBody(queries map[string]string, w io.Writer) error {

	t, err := template.New("body").Parse(body)
	if err != nil {
		return err
	}

	if err := t.Execute(w, queries); err != nil {
		return err
	}

	return nil
}

// SendReport takes a slice of strings and sends a report in plain text to given email
func SendReport(text string, queries map[string]string) error {

	// setup
	f, t, key := initVars()
	from := mail.NewEmail(name, f)
	subject := "Sendgrid.net URLs report"
	to := mail.NewEmail(name, t)
	client := sendgrid.NewSendClient(key)

	// body
	var buff bytes.Buffer
	if err := createHTMLBody(queries, &buff); err != nil {
		return err
	}
	content := buff.String()
	htmlContent := content

	// attachment
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	attachment := mail.NewAttachment()
	attachment.SetType("text/html")
	attachment.SetContent(encoded)
	attachment.SetFilename("report.html")
	message := mail.NewSingleEmail(from, subject, to, content, htmlContent)
	message.AddAttachment(attachment)

	// send report
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	}
	log.Printf("report sent, Status Code: %d", response.StatusCode)
	return nil
}
