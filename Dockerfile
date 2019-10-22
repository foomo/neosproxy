# -----------------------------------------------------------------------------
# Builder Base
# -----------------------------------------------------------------------------
FROM golang:1.13-alpine as base

RUN apk add --no-cache git upx \
  && rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/foomo/neosproxy



# -----------------------------------------------------------------------------
# Builder
# -----------------------------------------------------------------------------
FROM base as builder

COPY . ./

# Build the binary
RUN go mod vendor
RUN CGO_ENABLED=0 go build -o /go/bin/neosproxy cmd/neosproxy/main.go

# Compress the binary
RUN upx /go/bin/neosproxy



# -----------------------------------------------------------------------------
# Container
# -----------------------------------------------------------------------------
FROM alpine:latest

RUN apk add --no-cache \
    tzdata ca-certificates \
  && rm -rf /var/cache/apk/*

# Required for alpine image and golang
RUN echo "hosts: files dns" > /etc/nsswitch.conf

COPY --from=builder /go/bin/neosproxy /usr/local/bin/neosproxy

VOLUME ["/var/data/neosproxy"]

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/neosproxy"]
