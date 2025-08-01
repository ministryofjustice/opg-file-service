FROM golang:1.24.5-alpine

RUN go install gotest.tools/gotestsum@latest \
  && addgroup -S app && adduser -S -g app app

USER app
