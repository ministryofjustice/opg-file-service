FROM golang:1.24 AS build-env

ARG TARGETARCH
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main.go main.go
COPY cache cache
COPY dynamo dynamo
COPY handlers handlers
COPY internal internal
COPY middleware middleware
COPY session session
COPY storage storage
COPY zipper zipper

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/zipper

FROM alpine:3

RUN apk --update --no-cache add \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*
RUN apk --no-cache upgrade busybox libcrypto3 libssl3

COPY --from=build-env /go/bin/zipper /go/bin/zipper

RUN addgroup -S app && adduser -S -g app app \
    && chown app:app /go/bin/zipper
USER app

ENTRYPOINT ["/go/bin/zipper"]
