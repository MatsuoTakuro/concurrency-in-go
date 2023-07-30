//go:build main_colored
// +build main_colored

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Fork uint8

const (
	F0 Fork = iota
	F1
	F2
	F3
	F4
)

var hungers = 3
var eatTime = 1 * time.Second
var thinkTime = 1 * time.Second
var sleepTime = 1 * time.Second

var orderMutex sync.Mutex
var orderFinished []string

type Philosopher struct {
	name  string
	right Fork
	left  Fork
	color *color.Color // for colored output of the philosopher in the terminal.
}

func (p *Philosopher) dine(
	dined *sync.WaitGroup,
	forks map[Fork]*sync.Mutex,
	seated *sync.WaitGroup,
) {
	defer dined.Done()

	p.Printf("%s is seated at the table.\n", p.name)
	seated.Done()

	// Wait until everyone is seated.
	seated.Wait()

	// Have this philosopher eatTime and thinkTime hungers times.
	for i := hungers; i > 0; i-- {
		p.Printf("%s gets hungry.\n", p.name)

		// Get a lock on the left and right forks. We have to choose the lower numbered fork first in order
		// to avoid a logical race condition, which is not detected by the -race flag in tests; if we don't do this,
		// we have the potential for a deadlock, since two philosophers will wait endlessly for the same fork.
		// Note that the goroutine will block (pause) until it gets a lock on both the right and left forks.
		if p.left < p.right {
			forks[p.left].Lock()
			p.Printf("\t%s takes the left fork.\n", p.name)

			forks[p.right].Lock()
			p.Printf("\t%s takes the right fork.\n", p.name)
		} else {
			forks[p.right].Lock()
			p.Printf("\t%s takes the right fork.\n", p.name)
			forks[p.left].Lock()
			p.Printf("\t%s takes the left fork.\n", p.name)
		}
		p.Printf("\t%s has both forks and is eating.\n", p.name)
		time.Sleep(eatTime)

		// The philosopher starts to think, but does not drop the forks yet.
		p.Printf("\t%s is thinking.\n", p.name)
		time.Sleep(thinkTime)

		// Unlock the mutexes for both forks.
		forks[p.left].Unlock()
		forks[p.right].Unlock()

		p.Printf("\t%s put down the forks.\n", p.name)
	}

	p.color.Printf("%s is satisified.\n", p.name)
	p.color.Printf("%s left the table.\n", p.name)

	orderMutex.Lock()
	orderFinished = append(orderFinished, p.name)
	orderMutex.Unlock()
}

var printMutex sync.Mutex // for locking output to the terminal because the fatih/color package is not thread-safe.

// TODO: sometimes the output is not colored. Why?
func (p *Philosopher) Printf(format string, a ...interface{}) {
	printMutex.Lock()
	defer printMutex.Unlock()
	p.color.Printf(format, a...)
}

var philosophers = []Philosopher{
	{
		name:  "Plato",
		left:  F4,
		right: F0,
		color: color.New(color.FgRed),
	},
	{
		name:  "Socrates",
		left:  F0,
		right: F1,
		color: color.New(color.FgGreen),
	},
	{
		name:  "Aristotle",
		left:  F1,
		right: F2,
		color: color.New(color.FgYellow),
	},
	{
		name:  "Pascal",
		left:  F2,
		right: F3,
		color: color.New(color.FgBlue),
	},
	{
		name:  "Locke",
		left:  F3,
		right: F4,
		color: color.New(color.FgMagenta),
	},
}

func main() {
	// print out a welcome message
	fmt.Println("Dining Philosophers Problem")
	fmt.Println("---------------------------")
	fmt.Println("The table is empty.")

	time.Sleep(sleepTime)

	// start the meal
	dine()

	// print out finished message
	fmt.Println("The table is empty.")

	time.Sleep(sleepTime)
	fmt.Printf("Order finished: %s.\n", strings.Join(orderFinished, ", "))
}

func dine() {

	var allDined sync.WaitGroup
	allDined.Add(len(philosophers))

	// We want everyone to be allSeated before they start eating
	var allSeated sync.WaitGroup
	allSeated.Add(len(philosophers))

	var forks = make(map[Fork]*sync.Mutex)
	for i := range philosophers {
		forks[Fork(i)] = &sync.Mutex{}
	}

	// each philosopher starts to dine respcetively
	for _, p := range philosophers {
		p := p // capture range variable because it changes in each iteration
		/*
			The line p := p is used to create a new instance of the p variable that is local to each iteration of the loop.
			This is necessary because the go keyword starts a new goroutine, which runs concurrently with the loop.
			If you don't create a new instance of p, all the goroutines will share the same variable, and they might see it change as the loop iterates.
			This is a common gotcha in Go when using goroutines inside loops.
			If you don't capture the loop variable, you might see unexpected behavior
			because all the goroutines might end up using the same value (the last value that the loop variable was set to).
			By doing p := p, you ensure that each goroutine gets its own copy of the loop variable,
			which doesn't change when the loop iterates. This is often called "capturing" the loop variable.

			In Go, when you create a goroutine inside a loop and the goroutine uses a variable from the loop, it doesn't get its own copy of that variable.
			Instead, all goroutines share the same instance of the loop variable.
			This is because goroutines are not executed immediately when they are encountered.
			They are scheduled to run concurrently and they might not get a chance to run until after the loop has finished executing.
			By the time a goroutine runs, the loop variable might have been updated several times.
			So, if you don't capture the loop variable by creating a new, local instance of it in each iteration (with p := p),
			all goroutines will see the same value for the loop variable, which is its final value after the loop has finished executing.
		*/
		go p.dine(&allDined, forks, &allSeated)
	}

	allDined.Wait()
}
