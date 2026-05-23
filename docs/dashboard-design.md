#design #documentation #grafana #solar

# Dashboard Design: Solar Monitoring

**Project**: [[overview|Enphase Exporter]]
**Last Updated**: 2026-01-21
**Dashboards**: Solar Overview, Solar Analytics

## Design Philosophy

The solar monitoring system uses **two complementary dashboards** with distinct purposes:

| Dashboard | Purpose | Time Range | Target User |
|-----------|---------|------------|-------------|
| **Solar Overview** | "How is my system doing RIGHT NOW?" | Last 6 hours | Daily glance |
| **Solar Analytics** | "What patterns exist over time?" | Last 24 hours | Deep analysis |

### Key Principles

1. **No Duplication** - Each metric appears in ONE dashboard only
2. **Clear Purpose** - Dashboards have distinct, non-overlapping goals
3. **Logical Grouping** - Related panels are grouped in collapsible sections
4. **Configurable** - User-adjustable parameters (like electricity rate) via variables

---

## Dashboard Variables

Both dashboards share these template variables:

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `$datasource` | Data source | Prometheus | Select Prometheus instance |
| `$rate` | Textbox | `0.21` | Electricity rate ($/kWh) for savings calculations |

### Why `$rate` Variable?

Hardcoding electricity rates creates maintenance burden and doesn't account for:
- Different utility rates by region
- Time-of-use pricing
- Rate changes over time

Users can adjust `$rate` in the dashboard UI to match their actual electricity cost.

---

## Solar Overview Dashboard

**UID**: `solar-overview`
**Purpose**: At-a-glance operational status

### Section 1: System Status

Real-time system health indicators.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| System Health | Stat | Combined inverter count + production check | Overall health indicator |
| Current Production | Stat | `enphase_production_watts{device_type="eim"}` | Live watts being produced |
| Current Consumption | Stat | `enphase_consumption_watts{measurement_type="total-consumption"}` | Live watts being consumed |
| Grid Status | Stat | `enphase_net_watts` | Importing/Exporting status |
| Active Inverters | Stat | `count(enphase_inverter_watts)` | Inverter count with threshold coloring |

**System Health Logic**:
```promql
(count(enphase_inverter_watts) >= 30) +
(enphase_production_watts{device_type="eim"} > 0 or hour() < 7 or hour() > 19)
```
- Value 2 = Healthy (green) - all inverters online + producing (or nighttime)
- Value 1 = Degraded (yellow) - partial issue
- Value 0 = Issue (red) - major problem

### Section 2: Today's Summary

Daily energy statistics reset at midnight.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Today's Production | Stat | `enphase_production_wh_today{device_type="eim"} / 1000` | kWh produced today |
| Today's Consumption | Stat | `enphase_consumption_wh_today{measurement_type="total-consumption"} / 1000` | kWh consumed today |
| Today's Self-Consumed | Stat | Production - Exported | kWh used directly from solar |
| Today's Savings | Stat | Self-consumed × `$rate` | Money saved today |

### Section 3: Lifetime Achievement

Cumulative statistics showing long-term value. Moved here from Analytics because these are "achievement badges" - perfect for at-a-glance viewing.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Lifetime Production | Stat | `enphase_production_wh_lifetime{device_type="eim"} / 1000000` | Total MWh ever produced |
| Lifetime Savings | Stat | Lifetime self-consumed × `$rate` | Total money saved |
| CO2 Avoided | Stat | Lifetime kWh × 0.417 kg/kWh | Environmental impact |
| Trees Equivalent | Stat | CO2 avoided / 22 kg/tree/year | Relatable comparison |

### Section 4: Power Flow

Real-time visualization of energy movement.

| Panel | Type | Queries | Purpose |
|-------|------|---------|---------|
| Production vs Consumption | Time Series | Production + Consumption overlaid | See patterns and relationship |

### Section 5: Efficiency Snapshot

Quick efficiency indicators as gauges (0-100%).

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Solar Offset % | Gauge | (Production / Consumption) × 100 | How much consumption comes from solar |
| Self-Consumption % | Gauge | ((Production - Exported) / Production) × 100 | How much solar is used locally |
| Power Factor | Gauge | `enphase_power_factor` | Grid efficiency metric |

---

## Solar Analytics Dashboard

**UID**: `solar-analytics`
**Purpose**: Deep analysis and trend discovery

### Section 1: Period Summary

Context-aware statistics based on selected time range (uses `$__range`).

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Produced | Stat | `increase(...[$__range])` | Energy produced in period |
| Consumed | Stat | `increase(...[$__range])` | Energy consumed in period |
| Self-Consumed | Stat | Produced - Exported | Direct solar usage |
| Exported | Stat | `increase(enphase_energy_exported_wh...[$__range])` | Sent to grid |
| Imported | Stat | `increase(enphase_energy_imported_wh...[$__range])` | Drawn from grid |
| Net Balance | Stat | Exported - Imported | Grid exchange balance |

### Section 2: Energy Flow Analysis

Visualize where solar energy goes over time.

| Panel | Type | Queries | Purpose |
|-------|------|---------|---------|
| Energy Flow | Stacked Area | Self-consumed, Exported, Imported | Energy distribution over time |
| Energy Sources | Pie Chart | From Solar vs From Grid | Current power source breakdown |

### Section 3: Production Analysis

Production patterns and trends.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Production Pattern | Time Series | Hourly production | Intraday pattern |
| Daily Production | Bar Chart | Daily totals | Day-over-day comparison |

### Section 4: Consumption Analysis

Understanding consumption patterns.

| Panel | Type | Queries | Purpose |
|-------|------|---------|---------|
| Consumption Comparison | Time Series | Total vs Net consumption | See grid dependency |
| Self-Sufficiency Trend | Time Series | Self-sufficiency % over time | Efficiency tracking |

### Section 5: Financial Analysis

Cost savings visualization.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Daily Savings | Bar Chart | Daily self-consumed × `$rate` | Money saved per day |
| Cumulative Savings | Time Series | Running total of savings | Track progress toward goals |

### Section 6: Inverter Performance

Per-panel monitoring (expanded by default, not collapsed).

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Inverter Output | Time Series | `enphase_inverter_watts` by serial | Individual panel production |
| Inverter Details | Table | Serial, current W, max W, last report | Quick status check |

### Section 7: Grid Health

Technical grid metrics. Moved here from Overview since these are analytical rather than operational.

| Panel | Type | Query | Purpose |
|-------|------|-------|---------|
| Voltage by Phase | Time Series | `enphase_voltage_volts` | Monitor voltage stability |
| Frequency | Time Series | `enphase_frequency_hz` | Grid frequency (60Hz nominal) |
| Current by Phase | Time Series | `enphase_current_amps` | Current draw monitoring |

---

## Architectural Decisions

### ADR-001: Two Dashboards vs One

**Decision**: Separate Overview and Analytics dashboards

**Context**: Single dashboard became cluttered with both real-time stats and historical analysis panels.

**Consequences**:
- ✅ Clear purpose for each dashboard
- ✅ Faster load times (fewer panels)
- ✅ Better mobile experience
- ❌ Users must switch between dashboards

### ADR-002: Lifetime Stats in Overview

**Decision**: Move Lifetime Achievement section from Analytics to Overview

**Context**: Lifetime stats (total production, CO2 saved, trees equivalent) are "achievement badges" that users want to see at a glance, not analyze over time.

**Consequences**:
- ✅ Immediate visibility of long-term impact
- ✅ Motivational for daily viewing
- ✅ Analytics focuses purely on trends

### ADR-003: Grid Health in Analytics

**Decision**: Move Grid Health section from Overview to Analytics

**Context**: Voltage, frequency, and current by phase are technical metrics that don't need constant monitoring.

**Consequences**:
- ✅ Overview stays focused on solar operations
- ✅ Technical users can find detailed grid info in Analytics
- ✅ Reduces Overview clutter

### ADR-004: Configurable Electricity Rate

**Decision**: Add `$rate` variable instead of hardcoding `* 0.21`

**Context**: Electricity rates vary by region, utility, and time-of-use plans.

**Consequences**:
- ✅ Users can set their actual rate
- ✅ No code changes needed for rate updates
- ✅ Supports different rates for different users
- ❌ Users must remember to set the variable

---

## Metric Reference

### Production Metrics

| Metric | Labels | Unit | Description |
|--------|--------|------|-------------|
| `enphase_production_watts` | `device_type` | watts | Current production |
| `enphase_production_wh_today` | `device_type` | Wh | Today's production |
| `enphase_production_wh_total` | `device_type` | Wh | Lifetime production (counter) |
| `enphase_inverter_watts` | `serial` | watts | Per-inverter production |

### Consumption Metrics

| Metric | Labels | Unit | Description |
|--------|--------|------|-------------|
| `enphase_consumption_watts` | `measurement_type` | watts | Current consumption |
| `enphase_consumption_wh_today` | `measurement_type` | Wh | Today's consumption |
| `enphase_consumption_wh_total` | `measurement_type` | Wh | Lifetime consumption |

### Grid Metrics

| Metric | Labels | Unit | Description |
|--------|--------|------|-------------|
| `enphase_net_watts` | - | watts | Net grid flow (+import/-export) |
| `enphase_energy_exported_wh` | `phase` | Wh | Energy sent to grid |
| `enphase_energy_imported_wh` | `phase` | Wh | Energy from grid |
| `enphase_voltage_volts` | `phase` | volts | Grid voltage |
| `enphase_current_amps` | `phase` | amps | Grid current |
| `enphase_frequency_hz` | - | Hz | Grid frequency |
| `enphase_power_factor` | - | ratio | Power factor (0-1) |

### Important Label Values

- `device_type="eim"` - Energy Intelligence Module (system-level)
- `device_type="inverters"` - Aggregate inverter data
- `measurement_type="total-consumption"` - Total home consumption
- `measurement_type="net-consumption"` - Net (after solar offset)
- `phase="total"` - Combined phases
- `phase="phase-a"`, `phase="phase-b"` - Individual phases (split-phase)

---

## Files

| File | Location | Purpose |
|------|----------|---------|
| `dashboard-solar-overview.yaml` | `apps/enphase-exporter/` | Overview ConfigMap |
| `dashboard-solar-analytics.yaml` | `apps/enphase-exporter/` | Analytics ConfigMap |

Both deploy to `monitoring` namespace with label `grafana_dashboard: "1"` for sidecar pickup.

---

## Related

- [[overview]] - Project overview
- [[tasks/task-002-grafana-dashboards]] - Dashboard implementation task
- [[../flux/overview]] - Kubernetes cluster
