FROM golang:1.18.3

RUN go install gotest.tools/gotestsum@latest

