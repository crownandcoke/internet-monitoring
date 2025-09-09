# Internet Monitoring Stack v1.1

This is a complete, pre-configured monitoring stack for internet and network monitoring.

## Included Components
- Prometheus (metrics collection)
- Grafana (visualization) with custom dashboard
- Blackbox Exporter (endpoint monitoring)
- Speedtest Exporter v1.1 (internet speed testing)
- Node Exporter (system metrics)

## Quick Start
```bash
docker-compose up -d
```

## Access
- Grafana: http://localhost:3030 (admin/wonka)
- Prometheus: http://localhost:9090
- Speedtest: http://localhost:9696

## Configuration
All configurations are pre-set and ready to use. See DEPLOY.md for customization options.

## Docker Images Used
- crownandcoke/speedtest-exporter:v1.1
- prom/prometheus:latest
- grafana/grafana:latest
- prom/blackbox-exporter:latest
- prom/node-exporter:latest

Released: Tue, Sep  9, 2025 11:07:01 AM
