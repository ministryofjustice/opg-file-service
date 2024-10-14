FROM golang:1.23.1-alpine

RUN go install gotest.tools/gotestsum@latest

RUN addgroup -S app && adduser -S -g app app
USER app
