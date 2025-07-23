FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-api ./cmd/server/main.go

RUN useradd -r -u 10001 -g nogroup userapi


FROM scratch
#FROM ubuntu:22.04

COPY --from=builder /app/user-api /user-api

USER 10001

ENTRYPOINT ["/user-api"]