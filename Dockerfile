FROM golang:1.15.2-alpine3.12 AS build

WORKDIR /go/irma_bot

COPY db ./db
COPY go.mod .
COPY go.sum .
COPY main.go .
COPY storage ./storage
COPY telegram ./telegram
COPY cnf ./cnf

ENV CGO_ENABLED=0

RUN go test && go build -o /go/bin/irma_bot

FROM alpine:3.12

COPY --from=build /go/bin/irma_bot /usr/local/bin/irma_bot
COPY etc /etc/

RUN \
  adduser -DH user \
  \
  && apk add --no-cache \
    ca-certificates

USER user

ENV \
  IRMA_BOT_NAME= \
  IRMA_DB_ADDR= \
  IRMA_REDIS_ADDRS= \
  IRMA_TELEGRAM_PATH= \
  IRMA_TELEGRAM_PROXY= \
  IRMA_TELEGRAM_TOKEN= \
  IRMA_TELEGRAM_URL=

EXPOSE 8080

CMD ["/usr/local/bin/irma_bot"]
