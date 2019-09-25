package utils

import (
	"bytes"
	"fmt"
	"log"
	"mime/quotedprintable"
	"net/smtp"
)

// Inspired by https://github.com/tangingw/go_smtp

type EmailSender struct {
	Host string
	Port uint

	User     string
	Password string
}

func NewEmailSender(Host string, Port uint, Username, Password string) *EmailSender {
	if Host == "" {
		log.Fatal("missing smtp host")
	}
	if Port == 0 {
		log.Fatal("missing smtp port")
	}
	if Username == "" {
		log.Fatal("missing smtp username")
	}
	if Password == "" {
		log.Fatal("missing smtp password")
	}

	return &EmailSender{Host, Port, Username, Password}
}

func (sender *EmailSender) SendRawMail(to []string, msg string) error {

	err := smtp.SendMail(fmt.Sprintf("%s:%d", sender.Host, sender.Port),
		smtp.PlainAuth("", sender.User, sender.Password, sender.Host),
		sender.User, to, []byte(msg))

	if err != nil {
		return err
	}

	fmt.Printf("Mail sent successfully to %+v\n", to)
	return nil
}

func (sender *EmailSender) WriteEmail(dest []string, contentType, subject, bodyMessage string) string {

	header := make(map[string]string)
	header["From"] = sender.User

	receipient := ""

	for _, user := range dest {
		receipient = receipient + user
	}

	header["To"] = receipient
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", contentType)
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header["Content-Disposition"] = "inline"
	// Custom headers to disable some email providers links tracking
	header["trackopens"] = "false"
	header["trackclicks"] = "false"

	message := ""

	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	var encodedMessage bytes.Buffer

	finalMessage := quotedprintable.NewWriter(&encodedMessage)
	finalMessage.Write([]byte(bodyMessage))
	finalMessage.Close()

	message += "\r\n" + encodedMessage.String()

	return message
}

func (sender *EmailSender) WriteHTMLEmail(dest []string, subject, bodyMessage string) string {

	return sender.WriteEmail(dest, "text/html", subject, bodyMessage)
}

func (sender *EmailSender) WritePlainEmail(dest []string, subject, bodyMessage string) string {

	return sender.WriteEmail(dest, "text/plain", subject, bodyMessage)
}
