FROM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414

RUN go install gotest.tools/gotestsum@latest \
  && addgroup -S app && adduser -S -g app app

USER app
