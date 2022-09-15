FROM golang:1.19-alpine

WORKDIR /usr/src/l2plant-api

COPY go.mod go.sum ./

RUN go mod download

COPY src/ src/
COPY config/ config/
COPY config/ /config/
COPY main.go main.go
ENV CONFIG_DIR /config

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /usr/local/bin/l2planet-api main.go

CMD ["l2planet-api"]