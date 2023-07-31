package main

import (
	"fmt"
	"time"
)

func listenToChan(ch <-chan int) {
	for {
		i := <-ch // queue data into channel to the limit of the buffer
		fmt.Println("Got", i, "from channel")

		// simulate doing a lot of work
		time.Sleep(1 * time.Second)
	}
}

func main() {
	ch := make(chan int, 10)

	go listenToChan(ch)

	for i := 0; i <= 30; i++ {
		// the first 10 times through this loop, things go quickly; after that, things slow down.
		/*
			The speed of sending the first 10 items is fast because the channel buffer can hold up to 10 items before it blocks.
			However, the speed of receiving items from the channel is not affected by the buffer size; it still remains slow.

			In your code, the listenToChan goroutine receives an item from the channel and then sleeps for 1 second.
			This sleep operation is what slows down the receiving process.
			Regardless of how many items are in the channel buffer, the listenToChan goroutine will only process one item per second due to the sleep operation.

			So, while the sending of the first 10 items is fast, the receiving of those items is still slow because of the sleep operation in the listenToChan goroutine.
			After the first 10 items, the sending also becomes slow because the channel buffer is full and each send operation has to wait for
			an item to be received from the channel before it can proceed.


			There are a few ways to receive data from a buffered channel faster:

			1, Remove or reduce the sleep time:
				In your current code, the listenToChan goroutine sleeps for 1 second after receiving each item from the channel.
				This sleep operation is what's slowing down the receiving process.
				If you remove or reduce this sleep time, the goroutine will be able to receive items from the channel faster.

			2, Use multiple receiver goroutines:
				You can also speed up the receiving process by using multiple goroutines to receive data from the channel concurrently.
				Each goroutine would run the listenToChan function and they would all receive data from the channel at the same time.
				This is known as the "fan-out" pattern.
		*/
		fmt.Println("sending", i, "to channel...")
		ch <- i
		fmt.Println("sent", i, "to channel!")
	}

	fmt.Println("Done!")
	close(ch)
}
