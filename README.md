# Internet Monitoring Docker Stack

A comprehensive Docker stack for monitoring your home network performance, featuring real-time speed tests, ping monitoring, and beautiful Grafana dashboards. This stack includes Prometheus, Grafana, Blackbox Exporter, and a custom Go-based Speedtest Exporter.

<center><img src="images/image.png" width="4600" heighth="500"></center>

## Features

- 📊 **Real-time Network Monitoring** - Track ping times to multiple hosts
- 🚀 **Internet Speed Testing** - Automated and on-demand speed tests
- 📈 **Beautiful Dashboards** - Pre-configured Grafana dashboard with all metrics
- 🎯 **Custom Speedtest Exporter** - Efficient Go-based exporter with web UI
- ⏰ **Configurable Test Intervals** - Balance between monitoring and resource usage
- 🔧 **Manual Test Triggers** - Web interface for on-demand speed tests
- 🐳 **Docker Hub Integration** - Pre-built images available

## Prerequisites

- Docker and Docker Compose installed
- Port availability: 3030, 9090, 9115, 9696, 9100
- Internet connection for speed testing

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/crownandcoke/internet-monitoring
cd internet-monitoring
```

### 2. Configure Monitoring Targets (Optional)

Edit `prometheus/pinghosts.yaml` to customize ping monitoring targets:

```yaml
- targets:  # url;humanname;routing;switch
    - http://google.com;google.com;external;internetbox
    - http://github.com;github.com;external;internetbox  
    - http://reddit.com;reddit.com;external;internetbox
    - http://192.168.1.1;internetbox;local;internetbox
```

### 3. Start the Stack

```bash
docker-compose up -d
```

### 4. Access the Services

- **Grafana Dashboard**: http://localhost:3030 (login: admin/wonka)
  - Direct dashboard link: http://localhost:3030/d/o9mIe_Aik/internet-connection
- **Speedtest Control Panel**: http://localhost:9696
- **Prometheus**: http://localhost:9090
- **Blackbox Exporter**: http://localhost:9115

## Configuration

### Speed Test Intervals

By default, speed tests run every 2 hours to minimize resource usage. To adjust:

Edit `prometheus/prometheus.yml`:
```yaml
- job_name: 'speedtest'
  scrape_interval: 2h  # Change this value (e.g., 30m, 1h, 4h)
```

### Manual Speed Tests

Access the Speedtest Control Panel at http://localhost:9696 to:
- Trigger speed tests on-demand
- View current metrics
- Monitor test status

### Grafana Credentials

Default login: `admin` / `wonka`

**Important**: Change these credentials after first login via the Grafana settings panel.

## Components

### Services and Ports

| Service | Port | Description |
|---------|------|-------------|
| Grafana | 3030 | Metrics visualization dashboard |
| Prometheus | 9090 | Time-series database |
| Blackbox Exporter | 9115 | HTTP endpoint monitoring |
| Speedtest Exporter | 9696 | Internet speed testing with web UI |
| Node Exporter | 9100 | System metrics collection |

### Custom Speedtest Exporter

The stack includes a custom Go-based speedtest exporter (`crownandcoke/speedtest-exporter`) that:
- Provides efficient speed testing with minimal resource usage
- Includes a web UI for manual test triggers
- Caches results for 30 minutes
- Supports multiple test methods (Ookla CLI, speedtest-cli, HTTP fallback)
- Exposes Prometheus-compatible metrics

## Advanced Configuration

### Using Local Build Instead of Docker Hub

To build the speedtest exporter locally, edit `docker-compose.yml`:

```yaml
speedtest:
  # Comment out the image line
  # image: crownandcoke/speedtest-exporter:latest
  # Uncomment the build line
  build: ./speedtest-exporter
```

### Building and Updating Speedtest Exporter

```bash
cd speedtest-exporter
docker build -t crownandcoke/speedtest-exporter:latest .
# Optional: Push to Docker Hub (requires authentication)
docker push crownandcoke/speedtest-exporter:latest
```

## Resource Usage

- **CPU**: ~59% during speed test (5-10 seconds)
- **Memory**: ~15MB during tests, ~5MB idle
- **Network**: ~430MB down, ~160MB up per test
- **Frequency**: 12 tests per day with default 2-hour interval

## Monitoring URLs

- **Prometheus Targets**: http://localhost:9090/targets - View all monitored endpoints
- **Prometheus Graphs**: http://localhost:9090/graph - Query and visualize raw metrics
- **Blackbox Exporter**: http://localhost:9115 - View probe success/failure details
- **Speedtest Metrics**: http://localhost:9696/metrics - Raw speedtest metrics endpoint

## Troubleshooting

### No Data in Grafana

1. Check all services are running: `docker-compose ps`
2. Verify Prometheus targets are UP: http://localhost:9090/targets
3. Wait 5-10 minutes for initial data collection
4. Check container logs: `docker-compose logs [service_name]`

### Speed Test Not Working

1. Check speedtest container logs: `docker-compose logs speedtest`
2. Manually trigger a test at http://localhost:9696
3. Verify internet connectivity
4. Ensure ports 9696 is not blocked

### High CPU Usage

- Increase speedtest interval in `prometheus/prometheus.yml`
- Speed tests only use CPU during execution (5-10 seconds)
- Consider disabling automatic tests and using manual triggers only

## Docker Hub Images

The speedtest exporter is available on Docker Hub:
- Repository: `crownandcoke/speedtest-exporter`
- Tags: `latest`, `v1.1`, `v1.0`

### Version History
- **v1.1** (Latest)
  - Added monitoring for AI services (Claude, ChatGPT, APIs)
  - Improved dashboard with status history panel
  - Added logarithmic scale for HTTP duration metrics
  - Fixed gateway monitoring with ICMP support
  - Added http_2xx_4xx module for API monitoring
- **v1.0**
  - Initial release with Go-based speedtest exporter
  - Web UI for manual triggers
  - 30-minute result caching

## Contributing

Feel free to open issues or submit pull requests for improvements!

## Credits

- Original stack concept by [@vegasbrianc](https://github.com/vegasbrianc/github-monitoring)
- Custom speedtest exporter developed for multi-gigabit connection support
- Community contributions and feedback

## License

This project is open source and available under the MIT License.

## Security Note

This stack is designed for local network monitoring. If exposing to the internet:
- Change all default passwords
- Use a reverse proxy with authentication (e.g., Authelia)
- Configure firewall rules appropriately
- Consider using HTTPS for all services