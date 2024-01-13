package email

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"jacobo.tarrio.org/jtweb/comments/engine"
	"jacobo.tarrio.org/jtweb/comments/notification"

	mail "github.com/xhit/go-simple-mail/v2"
)

func NewEmailNotificationEngine(adminUri string, from string, to string, options ...NotificationEngineOption) notification.NotificationEngine {
	engine := &emailEngine{
		AdminUri: adminUri,
		Target:   mail.NewSMTPClient(),
		From:     from,
		To:       to,
	}
	SetHostPort("localhost:25")(engine)

	for _, option := range options {
		option(engine)
	}
	return engine
}

func SetHostPort(hostport string) NotificationEngineOption {
	return func(e *emailEngine) {
		e.ConnProvider = func() (net.Conn, error) {
			return net.Dial("tcp", hostport)
		}
	}
}

func SetUnixSocket(socket string) NotificationEngineOption {
	return func(e *emailEngine) {
		e.ConnProvider = func() (net.Conn, error) {
			return net.Dial("unix", socket)
		}
	}
}

func SetAuth(user string, pass string) NotificationEngineOption {
	return func(e *emailEngine) {
		e.Target.Username = user
		e.Target.Password = pass
	}
}

func SetAuthType(auth mail.AuthType) NotificationEngineOption {
	return func(e *emailEngine) {
		e.Target.Authentication = auth
	}
}

func SetEncryption(enc mail.Encryption) NotificationEngineOption {
	return func(e *emailEngine) {
		e.Target.Encryption = enc
	}
}

func AuthType(auth string) (mail.AuthType, error) {
	switch strings.ToUpper(auth) {
	case "":
		fallthrough
	case "AUTO":
		return mail.AuthAuto, nil
	case "NONE":
		return mail.AuthNone, nil
	case "PLAIN":
		return mail.AuthPlain, nil
	case "LOGIN":
		return mail.AuthLogin, nil
	case "CRAMMD5":
		fallthrough
	case "CRAM-MD5":
		return mail.AuthCRAMMD5, nil
	}
	return mail.AuthAuto, fmt.Errorf("unknown authentication type: %s", auth)
}

func Encryption(enc string) (mail.Encryption, error) {
	switch strings.ToUpper(enc) {
	case "STARTLS":
		return mail.EncryptionSTARTTLS, nil
	case "SSLTLS":
		return mail.EncryptionSSLTLS, nil
	case "":
		fallthrough
	case "NONE":
		return mail.EncryptionNone, nil
	}
	return mail.EncryptionNone, fmt.Errorf("unknown encryption type: %s", enc)
}

type NotificationEngineOption = func(*emailEngine)

type emailEngine struct {
	AdminUri     string
	Target       *mail.SMTPServer
	From         string
	To           string
	Auth         smtp.Auth
	ConnProvider func() (net.Conn, error)
}

// Notify implements notification.NotificationEngine.
func (e *emailEngine) Notify(comment *engine.Comment) error {
	email := mail.NewMSG().
		SetFrom(e.From).
		AddTo(e.To).
		SetSubject(fmt.Sprintf("New comment received from %s on %s", comment.Author, comment.PostId)).
		SetBody(mail.TextPlain, fmt.Sprintf(
			`A new comment was received.

Author: %[4]s
Text:
%[5]s

Approve: %[1]s#ApproveComment=approve,%[2]s,%[3]s
Reject: %[1]s#ApproveComment=reject,%[2]s,%[3]s
`, e.AdminUri, comment.PostId, comment.CommentId, comment.Author, comment.Text))

	if email.Error != nil {
		return email.Error
	}

	conn, err := e.ConnProvider()
	if err != nil {
		return err
	}

	target := *e.Target
	target.CustomConn = conn
	client, err := target.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	err = email.Send(client)
	if err != nil {
		return err
	}

	return nil
}
