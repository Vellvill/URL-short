FROM golang:1.17 as build

WORKDIR /app

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o metrics

EXPOSE 3535

ENTRYPOINT ["/app/metrics"]