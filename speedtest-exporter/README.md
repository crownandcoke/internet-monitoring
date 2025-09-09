# Speedtest Exporter for Prometheus

A lightweight Go-based Prometheus exporter that performs internet speed tests and exposes metrics compatible with the original stefanwalther/speedtest-exporter.

## Features

- Written in Go for minimal resource usage
- Compatible with existing Grafana dashboards
- Multiple speedtest backends:
  1. Ookla Speedtest CLI (most accurate)
  2. speedtest-cli (Python-based fallback)
  3. HTTP download test (always available)
- Caches results for 5 minutes to avoid excessive testing
- Automatic fallback if primary methods fail

## Metrics

The exporter provides the following Prometheus metrics:

- `speedtest_bits_per_second{direction="downstream"}` - Download speed in bits/second
- `speedtest_bits_per_second{direction="upstream"}` - Upload speed in bits/second
- `speedtest_ping` - Ping latency in milliseconds
- `up` - Exporter status (1 = running, 0 = error)

## Configuration

Environment variables:
- `SPEEDTEST_PORT` - Port to expose metrics on (default: 9696)

## Endpoints

- `/metrics` - Prometheus metrics endpoint (triggers test if cache expired)
- `/health` - Health check endpoint

## Building

```bash
docker build -t speedtest-exporter .
```

## Running

```bash
docker run -p 9696:9696 speedtest-exporter
```

The metrics will be available at http://localhost:9696/metrics