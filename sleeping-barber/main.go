package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

var seatingCapacity uint8 = 10
var arrivalRate uint8 = 100 // clients arriving at (roughly) regular intervals.
var cutTime = 1000 * time.Millisecond
var openTime = 10 * time.Second

func main() {
	rand.Seed(time.Now().UnixNano())

	color.Yellow("The Sleeping Barber Problem")
	color.Yellow("---------------------------")

	clientChan := make(chan string, seatingCapacity)
	doneChan := make(chan bool)

	shop := BarberShop{
		HairCutTime:     cutTime,
		NumberOfBarbers: 0,
		BarbersDoneChan: doneChan,
		ClientsChan:     clientChan,
		IsOpen:          false,
	}

	shop.addBarber("Frank")
	shop.addBarber("Gerard")
	shop.addBarber("Milton")
	shop.addBarber("Susan")
	shop.addBarber("Kelly")
	shop.addBarber("Pat")

	shopClosing := make(chan bool)
	closed := make(chan bool)

	// open the shop and close it after a certain amount of time
	go shop.Open(openTime, shopClosing, closed)

	var clientNum uint = 1
	go func() {
		for {
			randomMillsecNums := rand.Int() % (2 * int(arrivalRate))
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
