# Speedtest Exporter Observability Guide

This document describes the observability metrics available in the speedtest-exporter to detect blocking operations and monitor system health.

## Overview

The speedtest-exporter now includes comprehensive instrumentation to detect blocking operations, track performance, and diagnose issues. All metrics are exposed at the `/metrics` endpoint in Prometheus format.

## Key Metrics for Detecting Blocking Operations

### 1. HTTP Request Duration

**Metric:** `speedtest_http_request_duration_seconds`
**Type:** Histogram
**Labels:** `endpoint` (metrics, trigger, health, index)

This metric tracks how long each HTTP request takes to process. **This is the primary metric for detecting blocking operations.**

**What to look for:**
- Normal requests should complete in < 0.1 seconds
- Requests taking > 1 second indicate potential blocking
- Requests taking > 10 seconds indicate serious blocking issues

**Grafana Query Examples:**
```promql
# 95th percentile request duration
histogram_quantile(0.95, rate(speedtest_http_request_duration_seconds_bucket{endpoint="metrics"}[5m]))

# Requests taking longer than 1 second
histogram_quantile(0.99, rate(speedtest_http_request_duration_seconds_bucket[5m])) > 1

# Alert if /metrics endpoint is slow
rate(speedtest_http_request_duration_seconds_sum{endpoint="metrics"}[5m])
  /
rate(speedtest_http_request_duration_seconds_count{endpoint="metrics"}[5m]) > 1
```

### 2. Concurrent Requests

**Metric:** `speedtest_http_concurrent_requests`
**Type:** Gauge
**Labels:** `endpoint` (metrics, trigger, health, index)

Tracks the number of requests being processed simultaneously.

**What to look for:**
- Normal: 0-2 concurrent requests
- Warning: 3-5 concurrent requests (potential request pile-up)
- Critical: > 5 concurrent requests (blocking confirmed)

**Grafana Query:**
```promql
# Current concurrent requests
speedtest_http_concurrent_requests{endpoint="metrics"}

# Alert on request pile-up
speedtest_http_concurrent_requests{endpoint="metrics"} > 3
```

### 3. Speedtest Execution Duration

**Metric:** `speedtest_execution_duration_seconds`
**Type:** Histogram

Tracks how long the actual speed test takes to execute (not including HTTP overhead).

**What to look for:**
- Ookla speedtest: typically 10-30 seconds
- HTTP fallback: typically 5-10 seconds
- Timeouts occur at 60 seconds

**Grafana Query:**
```promql
# Average speedtest duration
rate(speedtest_execution_duration_seconds_sum[10m])
  /
rate(speedtest_execution_duration_seconds_count[10m])

# 99th percentile (detect slow tests)
histogram_quantile(0.99, rate(speedtest_execution_duration_seconds_bucket[10m]))
```

### 4. Speedtest Timeouts

**Metric:** `speedtest_timeouts_total`
**Type:** Counter
**Labels:** `method` (ookla, speedtest-cli, http-fallback)

Counts how many times each speedtest method has timed out.

**What to look for:**
- Zero timeouts is ideal
- Occasional timeouts (1-2/day) are acceptable
- Frequent timeouts indicate network or system issues

**Grafana Query:**
```promql
# Timeout rate per hour
rate(speedtest_timeouts_total[1h]) * 3600

# Total timeouts by method
sum by (method) (speedtest_timeouts_total)
```

### 5. Speedtest Skipped

**Metric:** `speedtest_skipped_total`
**Type:** Counter

Counts how many speedtest requests were skipped because a test was already running.

**What to look for:**
- Low skip rate (< 1/hour) is normal
- High skip rate indicates Prometheus is scraping too frequently
- Growing skip count suggests tests are taking longer than the scrape interval

**Grafana Query:**
```promql
# Skipped tests per hour
rate(speedtest_skipped_total[1h]) * 3600

# Alert if too many tests are being skipped
increase(speedtest_skipped_total[5m]) > 5
```

### 6. Speedtest Method Used

**Metric:** `speedtest_method_used_total`
**Type:** Counter
**Labels:** `method` (ookla, speedtest-cli, http-fallback)

Tracks which speedtest method was successfully used.

**What to look for:**
- Ideally: mostly "ookla" (most accurate)
- Warning: frequent "http-fallback" usage (speedtest tools failing)
- Investigate if method distribution changes unexpectedly

**Grafana Query:**
```promql
# Method distribution
sum by (method) (increase(speedtest_method_used_total[24h]))

# Percentage using fallback
sum(rate(speedtest_method_used_total{method="http-fallback"}[1h]))
  /
sum(rate(speedtest_method_used_total[1h])) * 100
```

## Enhanced Logging

The exporter now logs with timestamps and durations:

```
2025-10-20T20:30:00Z Starting speed test at 2025-10-20T20:30:00Z...
2025-10-20T20:30:15Z Speed test completed in 15.23 seconds - Download: 945.23 Mbps, Upload: 412.34 Mbps, Ping: 12.45 ms
```

**Special log messages:**
- `SLOW REQUEST: metrics took 2.34 seconds` - HTTP request took > 1 second (potential blocking)
- `Speed test already in progress, skipping...` - Concurrent test prevented
- `timed out after 60s` - Speedtest command hit timeout

## Recommended Grafana Dashboard Panels

### Panel 1: Request Duration Over Time
```promql
# Query
histogram_quantile(0.95, rate(speedtest_http_request_duration_seconds_bucket{endpoint="metrics"}[5m]))

# Alert threshold: > 1 second
```

### Panel 2: Concurrent Requests
```promql
# Query
speedtest_http_concurrent_requests

# Alert threshold: > 3
```

### Panel 3: Speedtest Duration
```promql
# Query
histogram_quantile(0.95, rate(speedtest_execution_duration_seconds_bucket[10m]))

# Info only, no alert needed
```

### Panel 4: Error Rates
```promql
# Timeout rate
rate(speedtest_timeouts_total[5m])

# Skip rate
rate(speedtest_skipped_total[5m])
```

### Panel 5: Method Success Distribution
```promql
# Query
sum by (method) (increase(speedtest_method_used_total[1h]))

# Visualization: Pie chart or bar graph
```

## Alerting Rules

Here are recommended Prometheus alerting rules:

```yaml
groups:
  - name: speedtest_blocking
    interval: 30s
    rules:
      - alert: SpeedtestHTTPBlocking
        expr: |
          rate(speedtest_http_request_duration_seconds_sum{endpoint="metrics"}[5m])
          /
          rate(speedtest_http_request_duration_seconds_count{endpoint="metrics"}[5m]) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Speedtest HTTP endpoint is blocking"
          description: "The /metrics endpoint is taking {{ $value | printf \"%.2f\" }} seconds on average, indicating blocking behavior."

      - alert: SpeedtestRequestPileup
        expr: speedtest_http_concurrent_requests{endpoint="metrics"} > 3
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Speedtest requests are piling up"
          description: "{{ $value }} concurrent requests detected, indicating severe blocking."

      - alert: SpeedtestHighTimeoutRate
        expr: rate(speedtest_timeouts_total[1h]) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High speedtest timeout rate"
          description: "Speedtests are timing out at {{ $value | printf \"%.2f\" }} per second."

      - alert: SpeedtestAlwaysSkipped
        expr: increase(speedtest_skipped_total[10m]) > 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Speedtests are being skipped"
          description: "{{ $value }} speedtests were skipped in 10 minutes, tests may be running too long."
```

## Troubleshooting with Metrics

### Scenario 1: System becomes unresponsive during speedtests

**Check these metrics:**
1. `speedtest_http_request_duration_seconds` - Are /metrics requests taking > 10s?
2. `speedtest_http_concurrent_requests` - Are requests piling up (> 5)?
3. `speedtest_execution_duration_seconds` - How long are tests actually taking?

**Expected behavior:**
- HTTP requests: < 0.1s (non-blocking design)
- Concurrent requests: 0-2
- Test duration: 10-30s

### Scenario 2: Prometheus shows gaps in data

**Check these metrics:**
1. `speedtest_skipped_total` - Are tests being skipped?
2. `speedtest_timeouts_total` - Are tests timing out?
3. Logs for "timed out" or "already in progress"

**Likely causes:**
- Tests taking longer than scrape interval
- Network connectivity issues
- System resource exhaustion

### Scenario 3: Unexpected speed test results

**Check these metrics:**
1. `speedtest_method_used_total` - Which method is being used?
2. `speedtest_timeouts_total` - Are preferred methods failing?

**Investigation:**
- If using "http-fallback" frequently → speedtest CLI tools may not be installed
- Check Docker logs for error messages

## Performance Baselines

Based on the fixes implemented:

| Metric | Before Fixes | After Fixes |
|--------|-------------|-------------|
| HTTP Request Duration (/metrics) | 5-10+ seconds | < 0.1 seconds |
| Concurrent Requests | 1-10+ | 0-2 |
| Speedtest Duration | 10-30 seconds | 10-30 seconds (unchanged) |
| CPU Usage During Test | 59% sustained | Brief spikes only |
| Timeout Rate | Unknown | < 0.01/hour |
| Skip Rate | Unknown | < 0.1/hour |

## Additional Resources

- Prometheus metrics best practices: https://prometheus.io/docs/practices/naming/
- Grafana dashboard examples: See `grafana/provisioning/dashboards/`
- Docker logs: `docker-compose logs -f speedtest`
