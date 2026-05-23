#project #status/active #go #prometheus #solar

# Project: Enphase Prometheus Exporter

**Status**: Active (Deployed)
**Priority**: Medium
**Created**: 2026-01-16
**Repository**: [github.com/rhwendt/enphase-exporter](https://github.com/rhwendt/enphase-exporter)
**Current Version**: v1.3.0

## Overview

Custom Prometheus exporter for Enphase IQ Gateway written in Go. Exposes solar production metrics to Grafana via the local gateway API.

**Why build custom?**
- Existing exporters (loafoe/prometheus-envoy-exporter) have bugs and limited metrics
- Full control over metrics and labels
- Local API = real-time, no cloud dependency

## Gateway Details

| Property | Value |
|----------|-------|
| mDNS Hostname | envoy.local |
| IP Address | 192.168.42.7 |
| Serial | 482530060013 |
| Firmware | D8.3.5286 |
| Model | IQ Gateway (X-IQ-AM1-240-5-HDK) |

## Core Metrics (v1)

| Metric | Description |
|--------|-------------|
| `enphase_production_watts` | Current solar production |
| `enphase_production_wh_total` | Lifetime production |
| `enphase_inverter_watts` | Per-panel production |
| `enphase_voltage_volts` | Grid voltage per phase |
| `enphase_current_amps` | Current per phase |
| `enphase_power_factor` | Power factor |
| `enphase_frequency_hz` | Grid frequency |

## Tech Stack

- **Language**: Go
- **Prometheus**: github.com/prometheus/client_golang
- **Config**: github.com/spf13/viper
- **Logging**: github.com/sirupsen/logrus
- **API Reference**: https://github.com/Matthew1471/Enphase-API

## Deployment

- **Namespace**: `enphase-exporter`
- **ArgoCD App**: `enphase-exporter`
- **Metrics Port**: 9090
- **Scrape Interval**: 60s (reduced from 30s to avoid stressing gateway)

### Kubernetes Resources
- Deployment (1 replica, revisionHistoryLimit: 3)
- Service (ClusterIP)
- ServiceMonitor (Prometheus)
- Secret (JWT token, gateway address)
- ConfigMaps for Grafana dashboards (in `monitoring` namespace)

## Grafana Dashboards

Two dashboards deployed as ConfigMaps with `grafana_dashboard: "1"` label. See [[dashboard-design]] for full documentation.

1. **Solar Overview** (`solar-overview`) - At-a-glance operational status
   - System Status: Health indicator, current production/consumption, grid status
   - Today's Summary: Production, consumption, self-consumed, savings
   - Lifetime Achievement: Total production, savings, CO2 avoided, trees equivalent
   - Power Flow: Production vs consumption time series
   - Efficiency Snapshot: Solar offset %, self-consumption %, power factor gauges

2. **Solar Analytics** (`solar-analytics`) - Deep analysis & trends
   - Period Summary: Produced, consumed, self-consumed, exported, imported, net balance
   - Energy Flow Analysis: Stacked area chart, energy sources pie chart
   - Production/Consumption Analysis: Patterns, daily totals, self-sufficiency trends
   - Financial Analysis: Daily savings, cumulative savings
   - Inverter Performance: Per-panel output, details table
   - Grid Health: Voltage, frequency, current by phase

### Dashboard Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `$datasource` | Prometheus | Data source selection |
| `$rate` | 0.21 | Electricity rate ($/kWh) for savings calculations |

## Known Issues & Fixes

### Split-Phase Lifetime Doubling (Fixed in v1.0.3)
The Enphase IQ Gateway firmware has a bug where top-level `whLifetime` values are doubled on split-phase systems. The exporter sums values from the `lines[]` array for lifetime metrics.

### Split-Phase Daily Reset Bug (Known Limitation)
The Enphase gateway has conflicting bugs on split-phase systems:
- **Top-level values** (`whToday`, `whLastSevenDays`) are **doubled**
- **Lines[] array values** do **NOT reset at midnight**

We use `lines[]` summation to avoid doubling, but this means daily/weekly consumption values may not reset properly at midnight. For accurate daily values, use the Enphase app (which uses the cloud API).

### Local API Limitations
- **No daily export/import metrics** - Only lifetime counters available for grid export/import. The Enphase app gets daily values from the cloud API.
- **Array grouping not available** - The local API doesn't expose which inverters belong to which array.

## Development Workflow

### 1. Track Work with GitHub Issues

**Always create a GitHub issue before starting work:**

```bash
# Create a bug report
gh issue create --title "Bug: consumption values doubled" --body "Description of the issue..."

# Create a feature request
gh issue create --title "Feature: add battery metrics" --body "Description..."

# List open issues
gh issue list
```

### 2. Create Feature Branch

**⚠️ NEVER commit directly to main!**

```bash
# Always start from latest main
git checkout main && git pull origin main

# Create feature branch (reference issue number if applicable)
git checkout -b fix/issue-123-consumption-doubled
# or
git checkout -b feat/battery-metrics
```

### 3. Local Development & Testing

**Before creating a PR, you MUST test locally against the real gateway.**

#### Option A: Quick Test (< 15 min)
Both production and local exporters can run simultaneously for short tests:

```bash
# Build and run locally
go build -o enphase-exporter ./cmd/exporter

export ENVOY_ADDRESS="https://192.168.42.7"
export ENVOY_SERIAL="482530060013"
export ENVOY_JWT="<your-jwt-token>"
./enphase-exporter
```

#### Option B: Extended Debugging (> 15 min)
Scale down production to avoid overwhelming the gateway:

```bash
# Scale down production exporter
kubectl scale deployment/enphase-exporter -n enphase-exporter --replicas=0

# Run local exporter and debug...
./enphase-exporter

# When done, scale back up
kubectl scale deployment/enphase-exporter -n enphase-exporter --replicas=1
```

#### Verify Metrics

```bash
# Check metrics output
curl -s http://localhost:9090/metrics | grep enphase_

# Compare key values against Enphase app:
# - production_wh_today should match app's "Produced Today"
# - consumption_wh_today should match app's "Consumed Today"
# - Check for doubled values, stale data, or anomalies
```

**Why this matters:** The Enphase gateway has known firmware bugs. Unit tests mock the API and won't catch these issues. Always verify metrics match the official Enphase app before releasing.

### 4. Create Pull Request

```bash
# Push branch
git push -u origin <branch-name>

# Create PR (references issue if applicable)
gh pr create --title "fix: resolve consumption doubling" --body "Fixes #123"
```

### 5. Merge & Release

1. Wait for CI checks to pass
2. Merge PR via GitHub (squash or merge)
3. **release-please** automatically creates a release PR
4. Merge release PR → Docker image builds and deploys

### 6. Deploy

ArgoCD auto-syncs, or manually trigger:

```bash
kubectl rollout restart deployment/enphase-exporter -n enphase-exporter
kubectl rollout status deployment/enphase-exporter -n enphase-exporter
```

## Related

- [[../flux/tasks/task-008-enphase-solar-monitoring]] - Original task (blocked on exporter bug)
- [[../flux/overview]] - Cluster infrastructure
- [[../apps/overview]] - ArgoCD apps

## Tasks

- [[tasks/task-001-implement-exporter]] - Main implementation (completed)
- [[tasks/task-002-grafana-dashboards]] - Dashboard improvements (completed)

## Documentation

- [[dashboard-design]] - Dashboard design, architecture decisions, and metric reference
