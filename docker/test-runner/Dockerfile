FROM golang:1.26-alpine@sha256:bd14630652464086289693533d25b791aa9ae7481e784d7eac5d4c948e9736ea

RUN go install gotest.tools/gotestsum@latest \
  && addgroup -S app && adduser -S -g app app

USER app
