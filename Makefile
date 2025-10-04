build:
	go build -o dist/api cmd/main.go

test:
	go test -v -race ./...

test-short:
	go test -short -v -race ./...

lint:
	golangci-lint run ./...

up:
	docker-compose up -d

down:
	docker-compose down --rmi local