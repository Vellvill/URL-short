FROM golang:1.17 as build

RUN mkdir /app
WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

ENV PORT 8080

WORKDIR ./cmd

RUN go build -o ./cmd

CMD ["./cmd"]
