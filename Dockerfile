FROM golang:latest

RUN go install github.com/cosmtrek/air@latest

WORKDIR /Users/andreyleonov/Desktop/goProject/server


COPY . .
RUN go mod tidy