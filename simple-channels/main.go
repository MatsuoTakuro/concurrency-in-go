package main

import (
	"fmt"
	"strings"
)

func shout(
	receiver <-chan string, // receive only channel
	sender chan<- string, // send only channel
) {
	for {
		// wait for something to be received on the ping channel and receive it
		s, ok := <-receiver
		if !ok {
			// if the channel is closed or empty, we break out of the loop
			// - It is closed if the channel is closed and no one is sending
			// - It is empty if the channel is unbuffered and no one is sending
			/*
				when the channel would be considered closed or empty.
				A channel is closed when no more values will be sent on it, which is indicated by ok being false.
				A channel is empty if it is unbuffered (it doesn't have a capacity specified) and no goroutine is currently sending a value on it.
				In this case, if a receive operation is attempted, it would block until a value is sent by another goroutine,
				unless the channel is closed, in which case the receive operation would return a zero value and ok would be false.
			*/
			break
		}

		// send the transformed text to the pong channel
		sender <- fmt.Sprintf("%s!!!", strings.ToUpper(s))
	}
}

func main() {
	ping := make(chan string)
	pong := make(chan string)

	// start a goroutine
	go shout(ping, pong)

	fmt.Println("Type something and press ENTER (enter Q to quit)")
	for {
		fmt.Print("-> ")

		var userInput string
		_, _ = fmt.Scanln(&userInput)

		// if the user enters "q" or "Q", we break out of the loop
		if userInput == strings.ToLower("q") {
			break
		}

		// send userInput to "ping" channel
		ping <- userInput

		// wait for a response from the pong channel and receive it
		resp := <-pong
		fmt.Println("response:", resp)
	}

	fmt.Println("All done. Closing channels.")
	// close the channels
	close(ping)
	close(pong)
}
