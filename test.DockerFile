FROM golang:1.19.3

RUN go install gotest.tools/gotestsum@latest

