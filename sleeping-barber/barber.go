package main

import (
	"sync"
	"time"

	"github.com/fatih/color"
)

type Barber struct {
	Name       string
	IsAsleep   bool
	sleepMutex sync.RWMutex
}

func NewBarber(name string) *Barber {
	return &Barber{
		Name:     name,
		IsAsleep: false,
	}
}

func (b *Barber) checkWaitingRoom() {
	color.Yellow("%s goes to the waiting room to check for clients.", b.Name)
}

func (b *Barber) comeToWork(shop *BarberShop) {
	color.Cyan("%s is coming to work.", b.Name)
	shop.makeBarberWork(b)
}

func (b *Barber) cutHair(client string) {
	color.Green("%s is cutting %s's hair.", b.Name, client)
	time.Sleep(CutTime)
	color.Green("%s is finished cutting %s's hair.", b.Name, client)
}

func (b *Barber) goHome() {
	color.Cyan("%s is going home.", b.Name)
}

func (b *Barber) sleep() {
	b.sleepMutex.Lock()
	defer b.sleepMutex.Unlock()
	color.Yellow("There is nothing to do, so %s takes a nap.", b.Name)
	b.IsAsleep = true
}

func (b *Barber) wakeUp() {
	b.sleepMutex.Lock()
	defer b.sleepMutex.Unlock()
	b.IsAsleep = false
}

func (b *Barber) isAsleep() bool {
	b.sleepMutex.RLock()
	defer b.sleepMutex.RUnlock()
	return b.IsAsleep
}
