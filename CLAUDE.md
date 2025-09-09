# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Docker-based network monitoring stack that uses Prometheus, Grafana, Blackbox Exporter, and a custom Speedtest Exporter to monitor home network connectivity and internet speeds.

## Key Commands

### Starting the Stack
```bash
docker-compose up -d
```

### Stopping the Stack
```bash
docker-compose down
```

### Viewing Logs
```bash
docker-compose logs -f [service_name]  # e.g., prometheus, grafana, blackbox_exporter, speedtest
```

### Rebuilding Speedtest Exporter (for local development)
```bash
cd speedtest-exporter
docker build -t crownandcoke/speedtest-exporter:latest .
docker push crownandcoke/speedtest-exporter:latest  # If updating Docker Hub
```

## Architecture

### Services and Ports
- **Prometheus** (port 9090): Time-series database that collects metrics from exporters
- **Grafana** (port 3030): Visualization dashboard (default credentials: admin/wonka)
- **Blackbox Exporter** (port 9115): Performs HTTP probes on configured targets
- **Speedtest Exporter** (port 9696): Custom Go-based exporter that runs internet speed tests
- **Node Exporter** (port 9100): Collects system metrics

### Custom Speedtest Exporter
- Written in Go for minimal resource usage
- Provides web UI at http://localhost:9696 for manual speed test triggers
- Exposes Prometheus metrics at /metrics endpoint
- Manual trigger endpoint at /trigger
- Caches results for 30 minutes to reduce load
- Uses multiple fallback test methods (Ookla CLI, speedtest-cli, HTTP tests)
- Docker Hub image: `crownandcoke/speedtest-exporter:latest`

## Key Configuration Files

- `prometheus/prometheus.yml`: Prometheus scrape configuration
  - Speedtest runs every 2 hours (scrape_interval: 2h)
  - Ping tests run every 5 seconds
- `prometheus/pinghosts.yaml`: Define hosts to monitor (format: `url;humanname;routing;switch`)
- `docker-compose.yml`: Service definitions and network configuration
  - Uses Docker Hub image for speedtest-exporter by default
  - Can switch to local build by uncommenting build directive
- `grafana/config.monitoring`: Grafana environment variables
- `grafana/provisioning/dashboards/ping-speed-stats.json`: Pre-configured dashboard
- `speedtest-exporter/`: Custom Go-based speedtest exporter source code
  - `main.go`: Core exporter logic with metrics endpoints
  - `index.html`: Web UI for manual speed test triggers
  - `Dockerfile`: Multi-stage build for minimal image size

## Important URLs

- Grafana Dashboard: http://localhost:3030/d/o9mIe_Aik/internet-connection
- Speedtest Control Panel: http://localhost:9696
- Prometheus Targets: http://localhost:9090/targets
- Blackbox Status: http://localhost:9115
- Speedtest Metrics: http://localhost:9696/metrics

## Development Notes

### Speedtest Exporter Updates
When modifying the speedtest-exporter:
1. Make changes to `speedtest-exporter/main.go` or `index.html`
2. Build locally: `cd speedtest-exporter && docker build -t crownandcoke/speedtest-exporter:latest .`
3. Test locally by switching docker-compose.yml to use build instead of image
4. Tag for versioning: `docker tag crownandcoke/speedtest-exporter:latest crownandcoke/speedtest-exporter:v1.x`
5. Push to Docker Hub: `docker push crownandcoke/speedtest-exporter:latest && docker push crownandcoke/speedtest-exporter:v1.x`

### Monitoring Configuration
- Ping hosts can be modified in `prometheus/pinghosts.yaml`
- Speedtest interval is set to 2 hours to reduce CPU/bandwidth usage
- Results are cached for 30 minutes
- Manual tests can be triggered via web UI at http://localhost:9696

### Resource Usage
- Speedtest uses ~59% CPU during test execution (5-10 seconds)
- Memory usage is minimal (~15MB during tests, ~5MB idle)
- Network usage per test: ~430MB download, ~160MB upload
- With 2-hour intervals: 12 tests per day

### Grafana Dashboard
- Dashboard is provisioned from `grafana/provisioning/dashboards/ping-speed-stats.json`
- Changes to dashboard should be exported and saved to this file
- Units are configured as "bps" for proper display of multi-gigabit speeds
- Includes link to speedtest control panel for manual triggers