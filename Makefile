build:
	go build -o dist/api cmd/main.go

test:
	go test -v -race ./...

test-short:
	go test -short -v -race ./...

lint:
	golangci-lint run ./...

build-docker:
	docker build . -t tabeo-myappts:local

up:
	docker-compose up -d

down:
	docker-compose down --rmi local

generate-sources:
	go generate ./...
	sqlc generate --file repository/sqlc.yaml