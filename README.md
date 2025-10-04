# Tabeo assignment

This repository contains the code for the Tabeo assignment. The project is structured as follows:

- `api/` contains api code definining http handlers, request/response DTOs and error mapping using the Chi library.
- `cmd/` contains the main application entry point along with the DI setup.
- `domain/` contains the core business logic and domain models including domain errors.
- `publichols` contains the public holidays api client with the logic to determine public holidays.
- `repository/` contains the repository layer for data persistence using sqlc for.
- `schema/` contains the database schema and migration files with golang-migrate tests written to check the migration
  works.

Other files:

- `Dockerfile` defines the Docker image for the application.
- `docker-compose.yml` sets up the application along with a PostgreSQL database.
- Makefile contains commands for building, running, and testing the application. (see below)

## Prerequisites

- Docker and Docker Compose installed on your machine.
- Go 1.25+ installed for local development.
- make installed for using the Makefile commands.
- sqlc installed for generating SQL code from the database schema.

## Unit test coverage

Tests are written simply using the Go testing package and testify for assertions. Any mocks have been written manually
to keep it simple, however, I've previously used `uber-mock` and `mockery`.
We write tests for every area of the application including:

- http controllers to test request validation and error response mapping - we do this by mapping the http.Handler
  defined in routes.go directly into the httptest.Server.
- domain logic remains isolated from the http layer and is tested directly.
- repository layer is tested using a test database spun up using `testcontainers-go` - this ensures that our sqlc
  generated code is also tested.
- public holidays api client is tested using a mock http server to simulate the external api.

## Setup and Running the Application

1. `make build-docker`
2. `make up`

## Running unit and integration tests

1. `make test` to run all tests
2. `make test-short` to exclude repository and (external) api tests

### Manual testing

```
POST /appts
{
"firstName": "John",
"lastName": "Doe",
"visitDate": "2026-01-03T00:00:00Z"
}
```

# Known issues and future improvements

Internal server errors leak into the output from the api - not good! We can mask this and log the output instead.
Appointment dates are modeled as time.Time and not scoped down to day - this could lead to 2 appointments on the same
day with different time. Must fix.
No e2e tests - considering publishing docs from Chi and using the generated docs to generate a test client - not sure
I'll get to this one :).