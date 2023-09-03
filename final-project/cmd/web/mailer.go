package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

const (
	EMAIL_HTML_TPML_NAME  = "email-html"
	EMAIL_PLAIN_TPML_NAME = "email-plain"
	EMAIL_BODY            = "body"
)

// Template is a template to render for the email body.
// It also represents the name of an email template.
type Template string

const (
	MAIL          Template = "mail"
	CONFIRM_EMAIL Template = "confirmation-email"
)

type EncryptType string

const (
	TLS  EncryptType = "tls"
	SSL  EncryptType = "ssl"
	NONE EncryptType = "none"
)

var EmailBaseTemplate = "./cmd/web/templates/%s.html.gohtml"

type Mailer struct {
	Domain        string
	Host          string
	Port          uint
	Username      string
	Password      string
	Encrypt       EncryptType
	FromAddress   string
	FromName      string
	AsyncMail     *sync.WaitGroup
	Msg           chan Message
	MailErr       chan error
	StopMail      chan bool
	AcceptMessage bool
	mutex         sync.RWMutex
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []*mail.File
	Data        any
	DataMap     map[string]any
	Template    Template
}

func (m *Mailer) sendMail(
	msg Message,
	errChan chan<- error, // send error
) {
	defer m.AsyncMail.Done() // decrement counter every time a message is sent

	if msg.Template == "" {
		msg.Template = MAIL
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if len(msg.DataMap) == 0 {
		msg.DataMap = make(map[string]any)
	}

	msg.DataMap["message"] = msg.Data

	htmlMsg, err := m.buildHTMLMessage(msg)
	if err != nil {
		errChan <- err
	}

	plainMsg, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errChan <- err
	}

	smtpServ := mail.NewSMTPClient()
	smtpServ.Host = m.Host
	smtpServ.Port = int(m.Port)
	smtpServ.Username = m.Username
	smtpServ.Password = m.Password
	smtpServ.Encryption = m.getEncryption()
	// close the connection after sending the email bacause we don't need to keep it open.
	// It also means we don't need to call Disconnect() later.
	smtpServ.KeepAlive = false
	smtpServ.ConnectTimeout = 10 * time.Second
	smtpServ.SendTimeout = 10 * time.Second

	smtpCli, err := smtpServ.Connect()
	if err != nil {
		errChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMsg)
	// AddAlternative allows you to add alternative parts to the body of the email message.
	// This is most commonly used to add an html version in addition to a plain text version that was already added with SetBody.
	email.AddAlternative(mail.TextHTML, htmlMsg)

	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			email.Attach(a)
		}
	}

	err = email.Send(smtpCli)
	if err != nil {
		errChan <- err
	}
	log.Printf("Email successfully sent To: %s, Subject: \"%s\", Message: \"%s\"\n", msg.To, msg.Subject, msg.Data)
}

func (m *Mailer) buildHTMLMessage(msg Message) (string, error) {
	baseTmpl := fmt.Sprintf(EmailBaseTemplate, msg.Template)

	htmlTmpl, err := template.New(EMAIL_HTML_TPML_NAME).ParseFiles(baseTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err = htmlTmpl.ExecuteTemplate(&b, EMAIL_BODY, msg.DataMap); err != nil {
		return "", err
	}

	htmlMsg := b.String()
	htmlCSSMsg, err := m.inlineCSS(htmlMsg)
	if err != nil {
		return "", err
	}

	return htmlCSSMsg, nil
}

func (m *Mailer) buildPlainTextMessage(msg Message) (string, error) {
	baseTmpl := fmt.Sprintf(EmailBaseTemplate, msg.Template)

	plainTmpl, err := template.New(EMAIL_PLAIN_TPML_NAME).ParseFiles(baseTmpl)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err = plainTmpl.ExecuteTemplate(&b, EMAIL_BODY, msg.DataMap); err != nil {
		return "", err
	}

	plainMsg := b.String()

	return plainMsg, nil
}

func (m *Mailer) inlineCSS(html string) (string, error) {
	opts := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	/*
		The code snippet you provided is configuring options for the premailer package, which is a Go library used to inline CSS styles into HTML documents.
		This is often done to ensure that HTML emails render consistently across different email clients, as not all email clients support linked or embedded stylesheets.

		Here's a brief explanation of the options you've listed:

		RemoveClasses: false:
			When set to true, this option would remove all class attributes from the HTML elements after inlining the CSS styles.
			By setting it to false, the class attributes are preserved in the HTML.
			By keeping class attributes in the HTML, it seems that you want to preserve the original structure of the HTML document.
			This could be useful if you have JavaScript that relies on these class attributes, or if you want to allow further styling or manipulation of the HTML after the inlining process.

		CssToAttributes: false:
			When set to true, this option would convert CSS properties into equivalent HTML attributes where possible (e.g., converting width in CSS to the width attribute in HTML).
			By setting it to false, this conversion is not performed.
			By not converting CSS properties to equivalent HTML attributes, you're likely aiming to keep the HTML as close to the original as possible.
			This might be important if you want to ensure that the HTML renders consistently across different environments,
			or if you have specific styling that can't be accurately represented using HTML attributes.

		KeepBangImportant: true:
			The !important declaration in CSS is used to give a CSS property higher importance than other rules.
			When this option is set to true, the !important declarations are kept in the inlined styles. If set to false, they would be removed.
			By keeping the !important declarations, you're likely trying to ensure that the specific styling rules that were marked as important in the original CSS are respected in the inlined version.
			This could be crucial if you have complex styling that relies on the !important declarations to render correctly.

		Overall, these settings suggest a desire to perform CSS inlining in a way that preserves the original structure and styling of the HTML as much as possible,
		without making significant alterations or simplifications.
		This could be important in scenarios where you have complex or nuanced styling, where you want to maintain compatibility with specific rendering environments,
		or where you want to allow further manipulation or processing of the HTML after the inlining process.
	*/

	prem, err := premailer.NewPremailerFromString(html, &opts)
	if err != nil {
		return "", err
	}

	htmlCSS, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return htmlCSS, nil
}

func (m *Mailer) getEncryption() mail.Encryption {
	switch m.Encrypt {
	case TLS:
		return mail.EncryptionSTARTTLS
	case SSL:
		return mail.EncryptionSSLTLS
	case NONE:
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
	/* NOTE: STARTTLS vs SSL/TLS
	The difference between EncryptionSSLTLS and EncryptionSTARTTLS lies in how they handle the encryption of the connection when sending email.
	Both are methods to secure the communication, but they work in slightly different ways:

	EncryptionSSLTLS (SSL/TLS):
		Connection: The connection is encrypted from the very beginning.
		Port: Typically uses port 465 for SMTP.
		Process: The client and server immediately negotiate the TLS (or SSL) encryption when the connection is made, before any email data is transmitted.
		Security: Since the connection is encrypted from the start, it's considered secure, but it might be considered less flexible than STARTTLS in some scenarios.

	EncryptionSTARTTLS (STARTTLS):
		Connection: The connection starts as plain text and then is upgraded to a secure connection.
		Port: Typically uses port 587 for SMTP.
		Process: The client connects to the server and then issues a STARTTLS command. The server responds, and they negotiate the TLS encryption.
							After that, the rest of the communication is encrypted.
		Security: Since the connection starts as plain text, there's a small window where a "man-in-the-middle" attack could theoretically occur.
							However, once the STARTTLS command is issued, the connection is encrypted and secure.

	In general, both methods are considered secure, and the choice between them might depend on the specific requirements of the email server
	you're communicating with or the preferences of your organization.

	Modern email systems often prefer STARTTLS because it's more flexible (it allows the option of encrypting or not encrypting) and
	because it uses a port (587) that's less likely to be blocked by firewalls. However, SSL/TLS is still widely supported and used.

	If you're configuring an email client or writing code to send email, you'll typically choose one of these options based on
	the requirements of the email server you're connecting to. Many email servers support both options,
	so you might also consider factors like compatibility with other systems, organizational policies, or specific security requirements.
	*/
}

func (m *Mailer) stopAcceptingMessage() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.AcceptMessage = false
	close(m.Msg)
}

func (m *Mailer) canAcceptMessage() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.AcceptMessage
}

func (m *Mailer) terminate() {
	close(m.MailErr)
	close(m.StopMail)
}
