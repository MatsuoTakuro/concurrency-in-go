package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MatsuoTakuro/final-project/data"
	"github.com/alexedwards/scs/v2"
)

type Server struct {
	Session  *scs.SessionManager
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	SentMail *sync.WaitGroup
	Models   data.Models
	Mailer   Mailer
}

func (s *Server) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: s.routes(),
	}

	s.InfoLog.Printf("Starting web server... at http://localhost:%s\n", WEB_PORT)
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func (s *Server) initMailer() {
	err := make(chan error)
	msg := make(chan Message, 100)
	stop := make(chan bool)

	s.Mailer = Mailer{
		Domain:        "localhost",
		Host:          "localhost",
		Port:          1025,
		Username:      "",
		Password:      "",
		Encrypt:       NONE,
		FromAddress:   "info@mycompany.com",
		FromName:      "Info",
		Send:          s.SentMail,
		MsgChan:       msg,
		ErrChan:       err,
		Stop:          stop,
		AcceptMessage: true,
		mutex:         sync.RWMutex{},
	}
	s.InfoLog.Printf("You may see sent mails at http://%s:%d\n", s.Mailer.Host, 8025)
}

func (s *Server) listenForMail() {
	for {
		/*
			The select statement in the listenForMail method will decide which case to block or unblock
			based on the availability of data on the channels associated with each case.
			If there is data available on more than one channel associated with the cases in the select statement,
			the select statement will choose one of the cases at random to execute.
		*/
		select {
		case msg := <-s.Mailer.MsgChan:
			/*
				There is data available on the s.Mailer.MsgChan channel if there is a message that can be received from the channel without blocking.
				This means that there is at least one message in the channel's buffer, or that there is a sender that is currently sending a message to the channel.
				If there is no data available on the s.Mailer.MsgChan channel, the select statement will block until data becomes available on the channel.
			*/
			go s.Mailer.sendMail(msg, s.Mailer.ErrChan)

		case err := <-s.Mailer.ErrChan:
			/*
				There is data available on the s.Mailer.ErrChan channel if there is an error that can be received from the channel without blocking.
				This means that there is at least one error in the channel's buffer, or that there is a sender that is currently sending an error to the channel.
				If there is no data available on the s.Mailer.ErrChan channel, the select statement will block until data becomes available on the channel.
			*/
			s.ErrorLog.Println(err)

		case <-s.Mailer.Stop:
			/*
				If a signal is received on the s.Mailer.Stop channel, the select statement will execute the logic in this case.
				If there is no signal on the s.Mailer.Stop channel, the select statement will block until a signal is received on the channel.
			*/
			s.InfoLog.Println("stopping sending mails...")
			return

			// default:
			// If there is no data available on any of the channels, the select statement will execute the logic in the default case.
		}
	}
}

func (s *Server) listenForShutdown() {
	quit := make(chan os.Signal, 1) // create a channel to receive signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // block until a signal is received
	s.shutdown()

	/* WARN: Use signal.NotifyContext along with context.Context in production instead!
		This way can be considered better than using signal.Notify with a manually created channel for several reasons:

	Simplification:
		By using signal.NotifyContext, you're leveraging the standard context package in Go, which can simplify your code.
		You don't need to create and manage a channel explicitly. The context package takes care of that for you.

	Integration with Other Contexts:
		If you're already using contexts elsewhere in your application (which is common in modern Go code), signal.NotifyContext integrates seamlessly with those contexts.
		You can create a child context from an existing context, allowing you to propagate cancellation or deadlines throughout your application.

	Resource Management:
		signal.NotifyContext returns a cancellation function that you should call to release resources when you're done with the context.
		This makes resource management more explicit and less error-prone.

	Idiomatic Code:
		Using contexts is considered idiomatic in modern Go code, especially when dealing with cancellation, timeouts, and passing request-scoped values.
		By using signal.NotifyContext, you're aligning your code with current best practices.

	Flexibility:
		Contexts provide a standardized way to handle cancellation and timeouts.
		By using a context, you can more easily extend your code in the future to handle additional cancellation conditions or to pass values in a request-scoped manner.

	There is a file named `shutdown_example.go` in the same directory that is just to show a simple example of how you could use signal.NotifyContext in your application.
	That way there is more concise and integrates better with other Go code that uses contexts.
	It's a more modern approach that leverages the powerful context package in Go, leading to cleaner and more maintainable code.

	Also, you can refer to another more realistic example for usage of signal.NotifyContext here -> (https://github.com/MatsuoTakuro/fcoin-balances-manager/blob/0da561455bcfcc3a54a9b6063a9e8c50e9e697dd/cmd/server.go#L30)
	*/
}

// shutdown gracefully shuts down the server
// closed is to receive a signal that mailer has closed
func (s *Server) shutdown() {
	// perform any cleanup tasks
	s.InfoLog.Println("would run cleanup tasks...")

	// stop accepting mails
	s.InfoLog.Println("stopping accepting new message to send...")
	s.Mailer.stopAcceptingMessage()

	// wait until all unsent mails queued in message channel are sent
	s.SentMail.Wait()
	s.Mailer.Stop <- true // send a signal to stop listening for mails after all mails are sent.
	// if you send the signal before all mails are sent, you may lose some mails.
	// because in the listenForMail method, the `select` may choose the `case` that is waiting for a signal to stop listening for mails at random to execute,
	// not `case` that is waiting for message to send via message channel that may not be empty after stopping accepting new messages to send.
	// Or it may be in the process of sending a message with the Mailer.sendMail method.

	s.InfoLog.Println("terminating mailer...")
	s.Mailer.terminate()

	s.InfoLog.Println("shutting down application...")
	os.Exit(0)
}
