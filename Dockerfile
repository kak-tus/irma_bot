FROM golang:1.12.2-alpine3.9 AS build

WORKDIR /go/irma_bot

COPY go.mod .
COPY go.sum .
COPY main.go .
COPY settings ./settings
COPY storage ./storage
COPY telegram ./telegram
COPY vendor ./vendor

RUN \
  apk add --no-cache \
    upx \
  \
  && go build -mod=vendor -o /go/bin/irma_bot \
  \
  && upx -9 /go/bin/irma_bot

FROM alpine:3.9

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
