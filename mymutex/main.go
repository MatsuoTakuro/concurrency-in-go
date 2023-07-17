package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

type Income struct {
	Source string
	Amount int
}

func main() {
	var bankBalance int
	var mutex sync.Mutex

	fmt.Printf("Initial account balance: $%d.00\n", bankBalance)

	incomes := []Income{
		{Source: "Main job", Amount: 500},
		{Source: "Gifts", Amount: 10},
		{Source: "Part time job", Amount: 50},
		{Source: "Investments", Amount: 100},
	}

	wg.Add(len(incomes))

	// loop through 52 weeks and print out how much is made; keep a running total
	for i, income := range incomes {
		go func(i int, income Income) {
			defer wg.Done()

			for week := 1; week <= 52; week++ {
				mutex.Lock()
				temp := bankBalance
				temp += income.Amount
				bankBalance = temp

				fmt.Printf("On week %d, you earned $%d.00 from %s and current balance is $%d.00\n", week, income.Amount, income.Source, bankBalance)
				mutex.Unlock()
			}
		}(i, income)
	}

	wg.Wait()

	fmt.Printf("Final bank balance: $%d.00\n", bankBalance)
}
