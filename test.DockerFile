FROM golang:1.19.1

RUN go install gotest.tools/gotestsum@latest

