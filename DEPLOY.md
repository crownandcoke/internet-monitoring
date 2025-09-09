# Internet Monitoring Stack v1.1 - Complete Deployment Guide

## What's Included

This complete monitoring stack includes:

### Services
- **Prometheus** - Metrics collection and storage
- **Grafana** - Visualization dashboards (custom configured)
- **Blackbox Exporter** - HTTP/HTTPS/ICMP endpoint monitoring
- **Speedtest Exporter v1.1** - Internet speed testing (crownandcoke/speedtest-exporter:v1.1)
- **Node Exporter** - System metrics

### Pre-configured Components

#### Grafana Dashboard (Updated in v1.1)
- Site Status panel with discrete measurement blocks
- HTTP Duration by Phase with logarithmic scale
- Total Response Time monitoring
- Speedtest gauges (Download/Upload/Ping)
- Manual speedtest trigger link

#### Blackbox Configuration (Updated in v1.1)
- `http_2xx` - Standard HTTP monitoring
- `http_2xx_4xx` - For API monitoring (accepts 403/401)
- `http_any` - Accepts any HTTP status
- `icmp` - Ping monitoring for gateways

#### Prometheus Configuration (Updated in v1.1)
- Pre-configured scrape jobs
- AI service monitoring (Claude, ChatGPT, APIs)
- Gateway monitoring via ICMP
- Dynamic target loading from pinghosts.yaml

## Quick Deploy

### Option 1: Clone and Run

```bash
# Clone the repository
git clone https://github.com/crownandcoke/internet-monitoring
cd internet-monitoring

# Start the stack
docker-compose up -d
```

### Option 2: Download Release Package

```bash
# Download the v1.1 release
wget https://github.com/crownandcoke/internet-monitoring/archive/refs/tags/v1.1.tar.gz
tar -xzf v1.1.tar.gz
cd internet-monitoring-1.1

# Start the stack
docker-compose up -d
```

## Configuration Files

All configuration files are included and pre-configured:

- `prometheus/prometheus.yml` - Scrape configurations
- `prometheus/pinghosts.yaml` - Monitoring targets
- `grafana/provisioning/dashboards/ping-speed-stats.json` - Dashboard
- `grafana/provisioning/datasources/datasource.yml` - Data source
- `blackbox/config/blackbox.yml` - Probe modules
- `docker-compose.yml` - Service definitions

## Docker Hub Images

- **Speedtest Exporter**: `crownandcoke/speedtest-exporter:v1.1`
- **Other services**: Using official images (prom/prometheus, grafana/grafana, prom/blackbox-exporter)

## Version 1.1 Changes

- Custom Go-based speedtest-exporter (replaces deprecated version)
- AI service monitoring (Claude.ai, ChatGPT, API endpoints)
- Enhanced Grafana dashboard with status history
- Gateway monitoring via ICMP
- Logarithmic scale for HTTP duration metrics
- New blackbox module for API monitoring (http_2xx_4xx)

## Access Points

- **Grafana**: http://localhost:3030 (admin/wonka)
- **Prometheus**: http://localhost:9090
- **Speedtest Control**: http://localhost:9696
- **Blackbox Exporter**: http://localhost:9115

## Customization

### Add Monitoring Targets

Edit `prometheus/pinghosts.yaml`:
```yaml
- targets:  # url;humanname;routing;switch;module
    - https://example.com;example;external;internetbox;http_2xx
```

### Change Speedtest Interval

Edit `prometheus/prometheus.yml`:
```yaml
- job_name: 'speedtest'
  scrape_interval: 30s  # Prometheus scrape (keep low)
  # Actual tests run on cache expiry (30 minutes)
```

## Support

- GitHub Issues: https://github.com/crownandcoke/internet-monitoring/issues
- Docker Hub: https://hub.docker.com/r/crownandcoke/speedtest-exporter