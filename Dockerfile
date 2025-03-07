FROM golang:1.22.8-alpine AS builder
RUN apk update && apk add build-base cmake gcc git
WORKDIR /go/src/github.com/u2u-labs/layerg-crawler
ADD . .
RUN go install
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
#RUN go build -ldflags -w
#RUN chmod +x layerg-crawler
WORKDIR /go/bin

FROM alpine:3.18
RUN apk add ca-certificates curl
# install migrate
#RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
#RUN mv migrate.linux-amd64 /usr/bin/migrate
ENV PATH="${PATH}:/go/bin"
WORKDIR /go/bin
COPY --from=builder /go/bin/* /go/bin/
ENTRYPOINT ["./layerg-crawler"]

