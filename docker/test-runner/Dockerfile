FROM golang:1.27-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc

RUN go install gotest.tools/gotestsum@latest \
  && addgroup -S app && adduser -S -g app app

USER app
