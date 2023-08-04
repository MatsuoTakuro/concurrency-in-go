package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

var SeatingCapacity uint8 = 10
var ArrivalRate uint8 = 100 // clients arriving at (roughly) regular intervals.
var CutTime = 1000 * time.Millisecond
var OpenTime = 10 * time.Second

func main() {
	rand.Seed(time.Now().UnixNano())

	color.Yellow("The Sleeping Barber Problem")
	color.Yellow("---------------------------")

	clientChan := make(chan string, SeatingCapacity)
	doneChan := make(chan bool)

	shop := NewBarberShop(doneChan, clientChan)

	NewBarber("Frank").comeToWork(shop)
	NewBarber("Gerard").comeToWork(shop)
	NewBarber("Milton").comeToWork(shop)
	NewBarber("Susan").comeToWork(shop)
	NewBarber("Kelly").comeToWork(shop)
	NewBarber("Pat").comeToWork(shop)

	shopClosing := make(chan bool)
	closed := make(chan bool)

	// open the shop and close it after a certain amount of time
	go shop.Open(OpenTime, shopClosing, closed)

	// accept clients at a certain rate
	var clientNum uint = 1
	go func() {
		for {
			randomMillsecNums := rand.Int() % (2 * int(ArrivalRate))
			select {
			case <-shopClosing:
				return
			case <-time.After(time.Millisecond * time.Duration(randomMillsecNums)):
				shop.accept(fmt.Sprintf("Client #%d", clientNum))
				clientNum++
			}
		}
	}()

	// wait for the shop to close
	<-closed
}
