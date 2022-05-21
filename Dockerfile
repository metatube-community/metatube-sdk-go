FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN apk add --update --no-cache --no-progress make git \
    && make server

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/javtube/javtube-sdk-go"

COPY --from=builder /src/build/javtube-server .

RUN apk add --update --no-cache --no-progress ca-certificates tzdata

ENV GIN_MODE=release
ENV PORT=8080
ENV TOKEN=""
ENV DSN=""
ENV AUTO_MIGRATE=1

ENTRYPOINT ["/javtube-server"]
