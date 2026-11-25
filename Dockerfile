# syntax=docker/dockerfile:1

# Run the builder on the native build platform and cross-compile for the
# target platform, instead of compiling under QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

RUN apk add --update --no-cache --no-progress make git

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . /src

ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH make server

FROM alpine:latest
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.source="https://github.com/metatube-community/metatube-sdk-go"

COPY --from=builder /src/build/metatube-server .

RUN apk add --update --no-cache --no-progress ca-certificates tzdata

ENV GIN_MODE=release
ENV PORT=8080
ENV TOKEN=""
ENV DSN=""
ENV REQUEST_TIMEOUT=""
ENV DB_MAX_IDLE_CONNS=0
ENV DB_MAX_OPEN_CONNS=0
ENV DB_PREPARED_STMT=0
ENV DB_AUTO_MIGRATE=0

EXPOSE 8080

ENTRYPOINT ["/metatube-server"]
