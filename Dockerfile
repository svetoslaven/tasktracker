# Build stage
FROM golang:1.22-alpine AS build

WORKDIR /tasktracker

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN CGO_ENABLED=0 go build -o tasktracker ./cmd/api

# Run stage
FROM alpine:latest

WORKDIR /tasktracker

COPY --from=build /tasktracker/tasktracker .

EXPOSE 8080

CMD ["./tasktracker"]
