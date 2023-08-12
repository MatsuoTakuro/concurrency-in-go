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

	"github.com/alexedwards/scs/v2"
)

type Server struct {
	Session  *scs.SessionManager
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Shutdown *sync.WaitGroup
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

	Here's a simple example of how you could use signal.NotifyContext in your application:

	```go
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	```

	This way is more concise and integrates better with other Go code that uses contexts.
	It's a more modern approach that leverages the powerful context package in Go, leading to cleaner and more maintainable code.
	You can refer to another more realistic example for usage of signal.NotifyContext here -> (https://github.com/MatsuoTakuro/fcoin-balances-manager/blob/0da561455bcfcc3a54a9b6063a9e8c50e9e697dd/cmd/server.go#L30)
	*/
}

func (s *Server) shutdown() {
	// perform any cleanup tasks
	s.InfoLog.Println("would run cleanup tasks...")

	// block until waitgroup is empty
	s.Shutdown.Wait()
	s.InfoLog.Println("closing channels and shutting down application...")
	os.Exit(0)
}
