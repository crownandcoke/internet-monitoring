# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod ./
COPY main.go .

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o speedtest-exporter .

# Final stage
FROM alpine:latest

# Install speedtest utilities and ping
RUN apk add --no-cache \
    ca-certificates \
    curl \
    iputils \
    py3-pip \
    && pip3 install speedtest-cli --break-system-packages \
    && rm -rf /var/cache/apk/*

# Try to install Ookla speedtest (optional, may not be available on all architectures)
RUN curl -s https://packagecloud.io/install/repositories/ookla/speedtest-cli/script.alpine.sh | sh || true \
    && apk add speedtest || true

WORKDIR /root/

COPY --from=builder /app/speedtest-exporter .
COPY index.html .

EXPOSE 9696

CMD ["./speedtest-exporter"]