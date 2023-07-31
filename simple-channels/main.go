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
		s := <-receiver

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
