FROM golang:1.19.0

RUN go install gotest.tools/gotestsum@latest

