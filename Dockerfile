FROM golang:1.17 as build

RUN mkdir /app
WORKDIR /app


COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o golang-short

EXPOSE 8080

ENTRYPOINT ["/short/cmd/main"]