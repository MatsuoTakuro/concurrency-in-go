package main

import (
	"fmt"
	"time"
)

func server1(ch chan<- string) {
	for {
		time.Sleep(2 * time.Second)
		ch <- "This is from server 1"
	}
}

func server2(ch chan<- string) {
	for {
		time.Sleep(3 * time.Second)
		ch <- "This is from server 2"
	}
}

func main() {
	fmt.Println("Select with channels")
	fmt.Println("--------------------")

	chan1 := make(chan string)
	chan2 := make(chan string)

	go server1(chan1)
	go server2(chan2)

	for {
		select {
		// because we have multiple cases listening to
		// the same channels, random ones are selected

		// the first case or the second case is selected randomly
		case s1 := <-chan1:
			fmt.Println("Case one:  ", s1)
		case s2 := <-chan1:
			fmt.Println("Case two:  ", s2)

		// the third case or the fourth case is selected randomly
		case s3 := <-chan2:
			fmt.Println("Case three:", s3)
		case s4 := <-chan2:
			fmt.Println("Case four: ", s4)

		default: // useful to avoid deadlock
			fmt.Println("No data received. Doing something else.")
			time.Sleep(1 * time.Second) // simulate doing something else
			/*
				This defaul case's behavior in the design of the select statement ensures that channel operations are given priority over the default case.
				This is useful in scenarios where you want to perform an operation if and only if no channel operations are ready,
				such as avoiding a deadlock or performing a default action when no data is available.

				However, avoiding a deadlock doesn't resolve a deadlock that has already occurred.
				Once a deadlock has occurred, it typically cannot be resolved programmatically within the context of the running program.
				The program is stuck in a state where it's waiting for resources to become available,
				but those resources are held by other parts of the program that are also waiting for resources.
				This is a circular dependency that can't be broken from within the program.

				The best way to handle deadlocks is to design your program to avoid them in the first place.
				This can involve careful use of synchronization primitives like mutexes and channels, and structuring your program in a way that avoids circular dependencies.
				For example, always acquiring locks in a consistent order can help avoid deadlocks.

				If a deadlock does occur, it typically requires external intervention to resolve, such as restarting the program.
				Tools like the Go race detector can help identify potential deadlocks and other concurrency-related issues during development.
			*/
		}
	}

}
