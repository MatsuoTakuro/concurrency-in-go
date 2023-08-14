package main

import (
	"sync"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	EncryptType string
	FromAddress string
	FromName    string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	Error       chan error
	Done        chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string // Template is a template to render for the email body.
}

func (m *Mail) sendMail(
	msg Message,
	errorChan chan<- error, // error channel to send errors
) {
	if msg.Template == "" {
		msg.Template = "mail"
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	data := map[string]any{
		"message": msg.Data,
	}
	msg.DataMap = data

	htmlMsg, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
	}

	plainMsg, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errorChan <- err
	}

	smtpServ := mail.NewSMTPClient()
	smtpServ.Host = m.Host
	smtpServ.Port = m.Port
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
		errorChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMsg)
	// AddAlternative allows you to add alternative parts to the body of the email message.
	// This is most commonly used to add an html version in addition to a plain text version that was already added with SetBody.
	email.AddAlternative(mail.TextHTML, htmlMsg)

	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			email.AddAttachment(a)
		}
	}

	err = email.Send(smtpCli)
	if err != nil {
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {

	return "", nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {

	return "", nil
}

func (m *Mail) getEncryption() mail.Encryption {
	switch m.EncryptType {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
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
