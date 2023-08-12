//go:build exclude

package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	srv := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulating some processing time
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "Hello, World!")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	// Create a context that will be canceled when a termination signal is received
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // Wait for termination signal

	// Create a context with a timeout for the graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	/* NOTE: why create a new context with a timeout for the graceful shutdown instead of using the existing context?
	The reason for creating a new context with a timeout specifically for the graceful shutdown,
	instead of using the existing context that was canceled by the termination signal,
	is to separate the concerns of signal handling and server shutdown.

	1. **Separation of Concerns**:
		The context created with `signal.NotifyContext` is specifically for listening to termination signals. Once a signal is received,
		this context is canceled. It doesn't have any additional information about how long the server should take to shut down or what should happen if the shutdown takes too long.

	2. **Specific Timeout for Shutdown**:
		By creating a new context with a timeout, you can specify exactly how long the server should take to shut down gracefully.
		This allows you to fine-tune the behavior of the shutdown process independently of the signal handling.

	3. **Avoiding Potential Conflicts**:
		If you were to add a timeout directly to the context used for signal handling, it could potentially conflict with the purpose of that context.
		For example, if you set a timeout on the signal handling context, it could expire before a signal is received, leading to unexpected behavior.

	4. **Clarity and Readability**:
		By using separate contexts for signal handling and server shutdown, the code is more clear and explicit about its intentions.
		Each context has a single, well-defined purpose, making the code easier to understand and maintain.

	Here's a breakdown of the code snippet:
	- The first context (`ctx`) is created to listen for termination signals. It will be canceled as soon as a termination signal is received.
	- The second context (`shutdownCtx`) is created with a specific timeout for the graceful shutdown process.
		It gives the server a fixed amount of time to shut down gracefully, allowing ongoing requests to complete.

	By using separate contexts for these two distinct purposes, the code is more robust, clear, and flexible.
	It allows you to handle signal detection and server shutdown in a way that's tailored to the specific needs and constraints of each task.
	*/
	defer cancel()

	// Shutdown the server with the context
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server Shutdown Failed:%+v", err)
	}

	fmt.Println("Server gracefully stopped")
}
