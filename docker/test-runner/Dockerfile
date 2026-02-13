FROM golang:1.25

RUN go install gotest.tools/gotestsum@latest \
  && addgroup -S app && adduser -S -g app app

USER app
