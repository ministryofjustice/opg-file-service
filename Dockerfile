FROM golang:1.14

RUN mkdir -p /app/src /app/bin

WORKDIR /app/src

COPY src/go.mod .
COPY src/go.sum .

RUN go mod download

COPY src /app/src

RUN go build -o /app/bin/zipper

ENTRYPOINT ["/app/bin/zipper"]
