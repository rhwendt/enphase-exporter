# Enphase Prometheus Exporter

A Prometheus exporter for Enphase IQ Gateway that exposes solar production metrics from your local gateway.

## Features

- **Local API access** - No cloud dependency, real-time metrics
- **Per-inverter metrics** - Monitor individual panel performance
- **Meter readings** - Voltage, current, power factor, frequency per phase
- **Production totals** - Current watts and lifetime kWh
- **Kubernetes ready** - Includes Deployment, Service, and ServiceMonitor manifests

## Metrics

| Metric | Description | Labels |
|--------|-------------|--------|
| `enphase_production_watts` | Current production in watts | `device_type` |
| `enphase_production_wh_total` | Lifetime production in Wh | `device_type` |
| `enphase_production_voltage_volts` | Grid voltage | `device_type` |
| `enphase_inverter_watts` | Per-inverter current production | `serial_number` |
| `enphase_inverter_max_watts` | Per-inverter max reported | `serial_number` |
| `enphase_active_power_watts` | Meter active power | `meter_id`, `phase` |
| `enphase_current_amps` | Meter current | `meter_id`, `phase` |
| `enphase_voltage_volts` | Meter voltage | `meter_id`, `phase` |
| `enphase_frequency_hz` | Grid frequency | `meter_id` |

## Quick Start

### 1. Generate JWT Token

Go to [entrez.enphaseenergy.com](https://entrez.enphaseenergy.com):

1. Log in with your Enlighten credentials
2. Search for your system by name
3. Select your gateway from the dropdown
4. Click "Create access token"
5. Copy the token (valid for 1 year)

### 2. Configure

```bash
cp .env.example .env
```

Edit `.env`:

```bash
ENVOY_ADDRESS=https://envoy.local   # Or use IP: https://192.168.x.x
ENVOY_SERIAL=your-gateway-serial
ENVOY_JWT=your-jwt-token
```

### 3. Run

**Local:**
```bash
go run ./cmd/exporter
```

**Docker:**
```bash
docker build -t enphase-exporter .
docker run -p 9090:9090 --env-file .env enphase-exporter
```

**Docker Compose:**
```bash
docker-compose up
```

### 4. Verify

```bash
curl http://localhost:9090/metrics | grep enphase_
```

## Kubernetes Deployment

```bash
# Create secret with your credentials
kubectl create secret generic enphase-exporter \
  --from-literal=serial=your-serial \
  --from-literal=jwt=your-jwt-token

# Deploy
kubectl apply -f deploy/kubernetes/
```

The ServiceMonitor will automatically configure Prometheus to scrape metrics.

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENVOY_ADDRESS` | Yes | - | Gateway URL (e.g., `https://envoy.local`) |
| `ENVOY_SERIAL` | Yes | - | Gateway serial number |
| `ENVOY_JWT` | Yes | - | JWT token from entrez.enphaseenergy.com |
| `EXPORTER_PORT` | No | `9090` | Metrics endpoint port |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | No | `text` | Log format (text, json) |

## Endpoints

| Path | Description |
|------|-------------|
| `/metrics` | Prometheus metrics |
| `/health` | Liveness probe (always returns 200) |
| `/ready` | Readiness probe (200 when authenticated) |

## Architecture

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────┐
│   Prometheus    │─────▶│ Enphase Exporter │─────▶│ IQ Gateway  │
│                 │scrape│                  │ API  │ (local)     │
└─────────────────┘      └──────────────────┘      └─────────────┘
                                │
                                ▼
                         ┌─────────────┐
                         │   Grafana   │
                         └─────────────┘
```

## Authentication Flow

```
1. User generates JWT at entrez.enphaseenergy.com (1 year validity)
2. Exporter validates JWT with gateway (/auth/check_jwt)
3. Gateway returns session cookie (30 min validity)
4. Exporter auto-refreshes session before expiry
5. API requests use session cookie
```

## Development

```bash
# Build
go build -o enphase-exporter ./cmd/exporter

# Build with version info
go build -ldflags "-X main.Version=1.0.0 -X main.GitCommit=$(git rev-parse HEAD)" ./cmd/exporter

# Run tests
go test ./...

# Lint
golangci-lint run
```

## License

MIT

## Acknowledgments

- [Enphase-API](https://github.com/Matthew1471/Enphase-API) - API documentation
- [go-envoy](https://github.com/loafoe/go-envoy) - Reference implementation
