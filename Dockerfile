FROM golang:1.17 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . ./
RUN go build -o app/cmd

EXPOSE 8080

CMD ["./golang-url-short"]
