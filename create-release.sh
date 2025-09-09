#!/bin/bash

# Create release package for Internet Monitoring Stack v1.1

VERSION="v1.1"
RELEASE_DIR="internet-monitoring-stack-${VERSION}"

echo "Creating release package for Internet Monitoring Stack ${VERSION}..."

# Create release directory
mkdir -p releases/${RELEASE_DIR}

# Copy all necessary files
cp docker-compose.yml releases/${RELEASE_DIR}/
cp -r prometheus releases/${RELEASE_DIR}/
cp -r grafana releases/${RELEASE_DIR}/
cp -r blackbox releases/${RELEASE_DIR}/
cp -r speedtest-exporter releases/${RELEASE_DIR}/
cp README.md releases/${RELEASE_DIR}/
cp DEPLOY.md releases/${RELEASE_DIR}/

# Create a release info file
cat > releases/${RELEASE_DIR}/RELEASE.md << EOF
# Internet Monitoring Stack ${VERSION}

This is a complete, pre-configured monitoring stack for internet and network monitoring.

## Included Components
- Prometheus (metrics collection)
- Grafana (visualization) with custom dashboard
- Blackbox Exporter (endpoint monitoring)
- Speedtest Exporter ${VERSION} (internet speed testing)
- Node Exporter (system metrics)

## Quick Start
\`\`\`bash
docker-compose up -d
\`\`\`

## Access
- Grafana: http://localhost:3030 (admin/wonka)
- Prometheus: http://localhost:9090
- Speedtest: http://localhost:9696

## Configuration
All configurations are pre-set and ready to use. See DEPLOY.md for customization options.

## Docker Images Used
- crownandcoke/speedtest-exporter:${VERSION}
- prom/prometheus:latest
- grafana/grafana:latest
- prom/blackbox-exporter:latest
- prom/node-exporter:latest

Released: $(date)
EOF

# Create tarball
cd releases
tar -czf internet-monitoring-stack-${VERSION}.tar.gz ${RELEASE_DIR}/
echo "Release package created: releases/internet-monitoring-stack-${VERSION}.tar.gz"

# Create checksum
sha256sum internet-monitoring-stack-${VERSION}.tar.gz > internet-monitoring-stack-${VERSION}.tar.gz.sha256
echo "Checksum created: releases/internet-monitoring-stack-${VERSION}.tar.gz.sha256"

echo "Release package ready for distribution!"