FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN go build -v -ldflags '-w -s -buildid=' -trimpath \
    -o javtube-server cmd/server/main.go

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/javtube/javtube-sdk-go"

RUN apk --no-cache --no-progress add ca-certificates tzdata

COPY --from=builder /src/javtube-server .

ENV GIN_MODE=release
ENV PORT=8080
ENV DSN=""
ENV AUTOMIGRATE=true

CMD ["/javtube-server"]
