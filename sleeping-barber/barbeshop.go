package main

import (
	"sync"
	"time"

	"github.com/fatih/color"
)

type BarberShop struct {
	HairCutTime     time.Duration
	NumberOfBarbers uint
	BarbersDoneChan chan bool
	ClientsChan     chan string
	IsOpen          bool
	mu              sync.RWMutex
}

// Open opens the barbershop for business.
// It will close the shop after a certain amount of time.
func (s *BarberShop) Open(openTime time.Duration, shopClosing, closed chan<- bool) {

	s.open()

	<-time.After(openTime)
	shopClosing <- true
	s.closeShopForDay()
	closed <- true
}

func (s *BarberShop) addBarber(barber string) {

	s.NumberOfBarbers++

	go func() {
		var isSleeping bool
		color.Yellow("%s goes to the waiting room to check for clients.", barber)

		for {
			// if there are no clients, the barber goes to sleep
			if len(s.ClientsChan) == 0 {
				color.Yellow("There is nothing to do, so %s takes a nap.", barber)
				isSleeping = true
			}

			// NOTE: The variable 'waitingForHaircut' remains true as long as there are clients still in the waiting room.
			// This holds true even when the shop is in the process of closing.
			client, waitingForHaircut := <-s.ClientsChan
			if waitingForHaircut {
				if isSleeping {
					color.Yellow("%s wakes %s up.", client, barber)
					isSleeping = false
				}
				// cut hair
				s.cutHair(barber, client)

			} else {
				// no clients left, so send the barber home
				// and close this goroutine
				s.sendBarberHome(barber)
				return
			}
		}
	}()
}

func (s *BarberShop) cutHair(barber, client string) {
	color.Green("%s is cutting %s's hair.", barber, client)
	time.Sleep(s.HairCutTime)
	color.Green("%s is finished cutting %s's hair.", barber, client)
}

func (s *BarberShop) sendBarberHome(barber string) {
	color.Cyan("%s is going home.", barber)
	s.BarbersDoneChan <- true
}

func (s *BarberShop) closeShopForDay() {
	color.Cyan("Closing shop for the day.")

	// accepting clients is now closed
	s.close()

	// wait for all barbers to finish a client's haircuts and go home
	for i := 1; i <= int(s.NumberOfBarbers); i++ {
		<-s.BarbersDoneChan
	}
	close(s.BarbersDoneChan)

	color.Green("---------------------------------------------------------------------")
	color.Green("The barbershop is now closed for the day, and everyone has gone home.")
}

func (s *BarberShop) accept(client string) {
	// print out a message
	color.Green("*** %s arrives!", client)

	if s.isOpen() {
		select {
		// if there is room in the buffer (seating capacity) of the channel, the cliet can be sent to the channel.
		// I mean, the client can seat in the waiting room inside the shop.
		case s.ClientsChan <- client:
			color.Yellow("%s takes a seat in the waiting room.", client)
		default:
			color.Red("The waiting room is full, so %s leaves.", client)
		}
	} else {
		color.Red("The shop is already closed, so %s leaves!", client)
	}
}

func (s *BarberShop) open() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsOpen = true
	color.Green("The shop is open for the day!")
}

func (s *BarberShop) close() {
	close(s.ClientsChan)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsOpen = false
}

func (s *BarberShop) isOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsOpen
}
