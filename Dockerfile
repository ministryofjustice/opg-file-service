FROM golang:1.14

RUN mkdir /app

ADD src /app

WORKDIR /app

RUN go mod download
