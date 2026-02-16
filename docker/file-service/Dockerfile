FROM golang:1.26 AS build-env

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 65532 \
    app

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
COPY storage storage
COPY zipper zipper

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/zipper

FROM scratch

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /etc/group /etc/group

COPY --from=build-env /go/bin/zipper /go/bin/zipper

USER app:app

ENTRYPOINT ["/go/bin/zipper"]
