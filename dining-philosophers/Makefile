.PHONY: run test

# go xxx -race enable data race detection
run:
	go run -tags main -race .

test:
	go test -tags main -race -v ./...

run_colored:
	go run -tags main_colored -race .
	