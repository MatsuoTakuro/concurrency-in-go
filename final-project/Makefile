BINARY_NAME=myapp
DSN="host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable timezone=UTC connect_timeout=5"
REDIS="127.0.0.1:6399"

## build: Build binary
build:
	@echo "Building..."
	env CGO_ENABLED=0  go build -ldflags="-s -w" -o ${BINARY_NAME} ./cmd/web
	@echo "Built!"

## run: builds and runs the application setting the DSN and REDIS env vars
run: build
	@echo "Starting..."
	env DSN=${DSN} REDIS=${REDIS} ./${BINARY_NAME} &
	@echo "Started!"

## ps: shows the running application process info
ps:
	@PID=$(shell pgrep -f './${BINARY_NAME}'); \
	if [ -n "$$PID" ]; then \
		ps -p $$PID -o pid,ppid,user,stat,comm,args,start,time,%cpu,%mem; \
	else \
		echo "Process './${BINARY_NAME}' not found."; \
	fi

## clean: runs go clean and deletes binaries
clean:
	@echo "Cleaning..."
	@go clean
	@rm ${BINARY_NAME}
	@echo "Cleaned!"

## start: an alias to run
start: run

## stop: stops the running application
stop:
	@echo "Stopping..."
	-pkill -SIGTERM -f "./${BINARY_NAME}"
	@echo "Stopped!"

## restart: stops and starts the application
restart: stop start

## test: runs all tests
test:
	go test -v ./... -count=1

coverage:
	go test ./... -coverprofile=./coverage.out
	go tool cover -html=coverage.out

## db_in: access the database
db_in:
	@echo "Access the database..."
	PGPASSWORD='password' psql -h localhost -p 5432 -U postgres -d concurrency
	@echo "Finished access to the database!"

## session_in: access the redis
session_in:
	@echo "Access the redis as session store..."
	redis-cli -h 127.0.0.1 -p 6399
	@echo "Finished access to the redis!"

## login: login to the application
login:
	curl -X POST -d "email=admin@example.com&password=yourpassword" http://localhost:80/login
