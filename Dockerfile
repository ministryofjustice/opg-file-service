FROM golang:1.17.3 as build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/zipper

FROM alpine:3.14.2

RUN apk --update --no-cache add \
    ca-certificates \
    && rm -rf /var/cache/apk/*
RUN apk --no-cache add tzdata

COPY --from=build-env /go/bin/zipper /go/bin/zipper

RUN addgroup -S app && adduser -S -g app app \
    && chown app:app /go/bin/zipper
USER app

ENTRYPOINT ["/go/bin/zipper"]
