FROM golang:1.14 as build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/zipper

FROM scratch
COPY --from=build-env /go/bin/zipper /go/bin/zipper
ENTRYPOINT ["/go/bin/zipper"]
