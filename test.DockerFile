FROM golang:1.18.5

RUN go install gotest.tools/gotestsum@latest

