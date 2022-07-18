FROM golang:1.18.4

RUN go install gotest.tools/gotestsum@latest

