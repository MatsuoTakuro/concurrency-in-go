package main

import (
	"sync"
	"time"

	"github.com/fatih/color"
)

type BarberShop struct {
	Barbers         []*Barber
	BarbersDoneChan chan bool
	ClientsChan     chan string
	IsOpen          bool
	openMutex       sync.RWMutex
}

func NewBarberShop(doneChan chan bool, clientChan chan string) *BarberShop {
	return &BarberShop{
		Barbers:         []*Barber{},
		BarbersDoneChan: doneChan,
		ClientsChan:     clientChan,
		IsOpen:          false,
	}
}

// Open opens the barbershop for business.
// It will close the shop after a certain amount of time.
func (s *BarberShop) Open(openTime time.Duration, shopClosing, closed chan<- bool) {

	s.open()

	// wait for the shop to close after a certain amount of time
	<-time.After(openTime)
	shopClosing <- true
	s.closeShopForDay()
	closed <- true
}

func (s *BarberShop) makeBarberWork(barber *Barber) {

	s.Barbers = append(s.Barbers, barber)

	go func() {
		// check the waiting room for clients
		barber.checkWaitingRoom()
		for {
			// if there are no clients, the barber goes to sleep
			if len(s.ClientsChan) == 0 {
				barber.sleep()
			}

			// NOTE: The variable 'waitingForHaircut' remains true as long as there are clients still in the waiting room.
			// This holds true even when the shop is in the process of closing.
			client, waitingForHaircut := <-s.ClientsChan
			if waitingForHaircut {
				if barber.isAsleep() {
					color.Yellow("%s wakes %s up.", client, barber)
					barber.wakeUp()
				}
				// cut hair
				s.makeBarberCutHair(barber, client)

			} else {
				// no clients left, so send the barber home
				// and close this goroutine
				s.sendBarberHome(barber)
				return
			}
		}
	}()
}

func (s *BarberShop) makeBarberCutHair(barber *Barber, client string) {
	barber.cutHair(client)
}

func (s *BarberShop) sendBarberHome(barber *Barber) {
	barber.goHome()
	s.BarbersDoneChan <- true
}

func (s *BarberShop) closeShopForDay() {
	color.Cyan("Closing shop for the day.")

	// accepting clients is now closed
	s.close()

	// wait for all barbers to finish a client's haircuts and go home
	for i := 1; i <= int(len(s.Barbers)); i++ {
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
	s.openMutex.Lock()
	defer s.openMutex.Unlock()
	s.IsOpen = true
	color.Green("The shop is open for the day!")
}

func (s *BarberShop) close() {
	close(s.ClientsChan)

	s.openMutex.Lock()
	defer s.openMutex.Unlock()
	s.IsOpen = false
}

func (s *BarberShop) isOpen() bool {
	s.openMutex.RLock()
	defer s.openMutex.RUnlock()
	return s.IsOpen
}
