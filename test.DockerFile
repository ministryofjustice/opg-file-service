FROM golang:1.19.2

RUN go install gotest.tools/gotestsum@latest

