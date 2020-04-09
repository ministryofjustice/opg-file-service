FROM 1.13.10-alpine3.11

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go get && go run main.go