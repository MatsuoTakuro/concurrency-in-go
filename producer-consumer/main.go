package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

const MAX_NUM_OF_MAKING_PIZZAS = 10

var pizzasMade, pizzasFailed, total uint

type Producer struct {
	order chan PizzaOrder
	quit  chan chan error // used to send quit signals to the producer's goroutines.
}

type PizzaOrder struct {
	pizzaNumber uint
	message     string
	success     bool
}

func (p *Producer) Quit() error {
	// this new quit channel is used to wait for a confirmation that the Start method has finished its work.
	quit := make(chan error)
	p.quit <- quit // send the channel to the quit channel
	return <-quit  // wait for the quit channel to be closed
}

// close closes the order and quit channels of the producer.
// It should be called when the producer is not needed anymore.
func (p *Producer) close(quit chan error) {
	// close order channel
	close(p.order)
	// close quit channel and allow Quit() to stop waiting
	close(quit)

	/*
		In general, in Go, you should only close a channel when you know that no more values will be sent on it.
		In this case, since the p.quit channel is used to send quit signals,
		and you might need to send more than one quit signal during the lifetime of your program,
		it's better not to close it.
	*/
	// WARN: do not close the quit channel of the producer here!
	// close(p.quit)
}

func (p *Producer) Start() {
	var current_num uint

	for {
		result := p.makePizza(current_num)
		if result != nil {
			current_num = result.pizzaNumber
			select {
			// we tried to make a pizza (we sent something to the dat channel) whether it was successful or not.
			case p.order <- *result:
				continue

			case quit := <-p.quit:
				p.close(quit)
				return
			}
		}
	}
}

func (p *Producer) makePizza(pizzaNumber uint) *PizzaOrder {
	pizzaNumber++

	if pizzaNumber <= MAX_NUM_OF_MAKING_PIZZAS {
		fmt.Printf("Received order #%d!\n", pizzaNumber)

		delay := rand.Intn(3) + 1 // random number between 1 and 3
		fmt.Printf("Making pizza #%d. It will take %d seconds...\n", pizzaNumber, delay)
		time.Sleep(time.Duration(delay) * time.Second)

		rnd := rand.Intn(12) + 1 // random number between 1 and 12
		var msg string
		var success bool

		if rnd < 5 {
			pizzasFailed++
			if rnd == 2 {
				msg = fmt.Sprintf("*** We ran out of ingredients for pizza #%d!", pizzaNumber)
			} else {
				msg = fmt.Sprintf("*** The cook quit while making pizza #%d!", pizzaNumber)
			}
		} else {
			pizzasMade++
			success = true
			msg = fmt.Sprintf("Pizza order #%d is ready!", pizzaNumber)
		}
		total++

		p := PizzaOrder{
			pizzaNumber: pizzaNumber,
			message:     msg,
			success:     success,
		}

		return &p
	}

	// no more pizzas to make
	return &PizzaOrder{
		pizzaNumber: pizzaNumber,
	}
}

func main() {
	rand.Seed(time.Now().UnixNano()) // seed the random number generator

	color.Cyan("The Pizzeria is open for business!")
	color.Cyan("----------------------------------")

	pizzaJob := &Producer{
		order: make(chan PizzaOrder),
		quit:  make(chan chan error),
	}

	// start the producer to make pizzas
	go pizzaJob.Start()
	/*
		If you call pizzaJob.Start() as a goroutine twice, you'll have two goroutines running the Start method concurrently.
		This could lead to several issues:

		Race conditions: Both goroutines will be accessing and modifying the same Producer object (pizzaJob),
		which could lead to race conditions. For example, they might both try to increment current_num at the same time, leading to inconsistent results.

		Double closing of channels: If one goroutine receives a quit signal and closes the order channel,
		and then the other goroutine also tries to close the order channel, this will cause a panic, because closing an already closed channel in Go is a runtime error.

		Confusing output: The output of your program could be confusing, because the two goroutines might be interleaving their print statements.
		For example, you might see "Received order #1!" printed twice before seeing "Making pizza #1. It will take X seconds..." printed at all.

		In general, if you have a method that modifies shared state (like Start does), it's usually not safe to call it from multiple goroutines at the same time,
		unless the method has been specifically designed to be safe for concurrent use (for example, by using locks to protect the shared state).
		In your case, Start has not been designed for concurrent use, so it's best to only call it from one goroutine at a time.
	*/

	// consume pizza orders
	for result := range pizzaJob.order {
		if result.pizzaNumber <= MAX_NUM_OF_MAKING_PIZZAS {
			if result.success {
				color.Green(result.message)
				color.Green("Order #%d is out for delivery!", result.pizzaNumber)
			} else {
				color.Red(result.message)
				color.Red("Looks like the customer is really mad!")
			}
		} else {
			color.Cyan("Sorry, we are not taking any more orders.")
			err := pizzaJob.Quit()
			if err != nil {
				color.Red("*** Error closing channel!", err)
			}
		}
	}

	color.Cyan("-----------------")
	color.Cyan("Done for the day.")
	color.Cyan("We made %d pizzas, but failed to make %d, with %d attempts in total.", pizzasMade, pizzasFailed, total)
	switch {
	case pizzasFailed > 9:
		color.Red("It was an awful day...")
	case pizzasFailed >= 6:
		color.Red("It was not a very good day...")
	case pizzasFailed >= 4:
		color.Yellow("It was an okay day....")
	case pizzasFailed >= 2:
		color.Yellow("It was a pretty good day!")
	default:
		color.Green("It was a great day!")
	}
}
