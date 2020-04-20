FROM golang:1.14

RUN mkdir -p /app/src /app/bin

COPY src /app/src

WORKDIR /app/src

RUN go mod download
RUN go build -o /app/bin/zipper

ENTRYPOINT ["/app/bin/zipper"]
