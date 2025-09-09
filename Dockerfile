# Multi-service Internet Monitoring Stack
FROM alpine:latest

# Install docker-compose and required tools
RUN apk add --no-cache \
    docker-compose \
    curl \
    git

# Create working directory
WORKDIR /monitoring

# Copy the entire stack configuration
COPY docker-compose.yml .
COPY prometheus/ ./prometheus/
COPY grafana/ ./grafana/
COPY blackbox/ ./blackbox/
COPY speedtest-exporter/ ./speedtest-exporter/

# Add startup script
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'cd /monitoring && docker-compose up' >> /start.sh && \
    chmod +x /start.sh

# Document exposed ports
EXPOSE 3030 9090 9115 9696 9100

# Set labels
LABEL maintainer="crownandcoke" \
      version="1.1" \
      description="Complete Internet Monitoring Stack with Prometheus, Grafana, Blackbox, and Speedtest"

CMD ["/start.sh"]