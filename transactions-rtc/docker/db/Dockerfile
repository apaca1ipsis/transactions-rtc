# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.19 AS build-stage

WORKDIR /usr/local/go/src/transactions_rtc

COPY transactions_rtc/go.mod transactions_rtc/go.sum ./
RUN go get github.com/jackc/pgx/v5@latest
RUN go mod download

COPY transactions_rtc ./
COPY *.sql *.go ./

RUN cd db && CGO_ENABLED=0 GOOS=linux go build -o /db

# Run the tests in the container
FROM build-stage AS run-test-stage
# RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /db /db

USER nonroot:nonroot

ENTRYPOINT ["/db"]
