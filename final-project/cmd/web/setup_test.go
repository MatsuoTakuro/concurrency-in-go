package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MatsuoTakuro/final-project/data"
	"github.com/alexedwards/scs/v2"
)

var testServer Server

func TestMain(m *testing.M) {
	gob.Register(data.User{})

	testSession := scs.New()
	testSession.Lifetime = 24 * time.Hour
	testSession.Cookie.Persist = true
	testSession.Cookie.SameSite = http.SameSiteLaxMode
	testSession.Cookie.Secure = true

	testServer = Server{
		Session:   testSession,
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		AsyncJob:  &sync.WaitGroup{},
		AsyncErr:  make(chan error),
		StopAsync: make(chan bool),
	}

	// create a dummy mailer
	mailErr := make(chan error)
	msg := make(chan Message, 100)
	stopMail := make(chan bool)
	testServer.Mailer = Mailer{
		AsyncMail: testServer.AsyncJob,
		Msg:       msg,
		MailErr:   mailErr,
		StopMail:  stopMail,
	}

	// run a dummy mailer to listen for messages
	go func() {
		select {
		case <-testServer.Mailer.Msg:

		case err := <-testServer.Mailer.MailErr:
			testServer.InfoLog.Println(fmt.Errorf(ERROR_SENDING_MAIL_MSG, err))

		case <-testServer.Mailer.StopMail:
			testServer.InfoLog.Println("stopping sending mails...")
			return
		}
	}()

	// run a dummy listener for async job results
	go func() {
		select {
		case err := <-testServer.AsyncErr:
			testServer.ErrorLog.Println(fmt.Errorf(ERROR_ASYNC_JOB_MSG, err))

		case <-testServer.StopAsync:
			testServer.InfoLog.Println("stopping listening for errors...")
			return
		}
	}()

	os.Exit(m.Run())
}

// newReqWithSession return a request with loaded session in the context
// without this func, when you try to put some data in context without session, you will get an error
func newReqWithSession(rawReq *http.Request) *http.Request {
	ctxWithSession, err := testServer.Session.Load(rawReq.Context(), rawReq.Header.Get("X-Session")) // token as key for the session value is in header named X-Session
	/*
		In the context of web applications and HTTP requests, "loading a session" typically means retrieving session data from some storage mechanism
		and making it available for the current request. Sessions are used to maintain state between different HTTP requests from the same client.

		Here's a bit more detail on what "loading a session" might involve:
		Retrieval:
			The session data is usually stored either in-memory, in a database, or some other storage mechanism.
			The Load() function would retrieve this data based on a session identifier.

		Decoding:
			If the session data is stored in an encoded or encrypted format, the Load() function would decode or decrypt it.

		Populating Context
			Once the session data is retrieved and decoded, it's often added to the request context
			so that it can be easily accessed by other parts of the application that handle the request.

		Session Token:
			In your specific code snippet, the session identifier is expected to be in the X-Session header of the request.
			This is used to look up the session data.
	*/
	if err != nil {
		log.Println(err)
	}
	return rawReq.WithContext(ctxWithSession)
}
