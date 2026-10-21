# Contracts — Shared Data Contracts between Local EMS ↔ Cloud

This directory is the **single source of truth** for all data formats exchanged between `ems-local` (Go) and `ems-cloud` (NestJS).

## MQTT Topic Convention

| Topic Pattern | Direction | Description |
|---|---|---|
| `ems/{site_id}/telemetry` | Local → Cloud | Aggregated telemetry (interval is configurable; demo uses 10 sec) |
| `ems/{site_id}/heartbeat` | Local → Cloud | Heartbeat signal (every 30 sec) |
| `ems/{site_id}/alerts` | Local → Cloud | Real-time alerts |
| `ems/{site_id}/command` | Cloud → Local | Remote configuration commands |
| `ems/{site_id}/command/ack` | Local → Cloud | Command acknowledgment |
| `ems/{site_id}/failover/control` | Cloud → Local | Takeover and handover lease commands |
| `ems/{site_id}/failover/recovery` | Local → Cloud | Recovery request |
| `ems/{site_id}/failover/ack` | Bidirectional | Transition acknowledgment |

- `{site_id}` is a lowercase slug, e.g. `sg-01`, `tw-factory-02`.
- QoS 1 for telemetry, alerts, commands and acknowledgments; QoS 0 for heartbeat.
- Every message carries `schema_version: 1`; telemetry additionally carries a stable `message_id` for QoS 1 deduplication.

## Schema Files

| File | Used by | Description |
|---|---|---|
| [`telemetry.schema.json`](telemetry.schema.json) | Go publisher, NestJS consumer | Aggregated power, SOC, energy data |
| [`commands.schema.json`](commands.schema.json) | NestJS publisher, Go consumer | Remote config, tariff updates |
| [`heartbeat-v2.schema.json`](heartbeat-v2.schema.json) | Go publisher, NestJS consumer | Control owner, failover state, lease epoch and EV state |
| [`failover-v2.schema.json`](failover-v2.schema.json) | Go/NestJS | Monotonic takeover and handover coordination |

Cloud accepts heartbeat v1 during migration. Failover receivers ignore commands whose
`lease_epoch` is lower than the latest accepted epoch.

## Rules for Changing Contracts

1. **Edit the schema file first** — this is the source of truth.
2. **Update both sides in the same PR** — Go struct + TypeScript type.
3. **Never remove a field** — mark as deprecated, keep for backward compat.
4. **Add new fields as optional** — with `"default"` value in schema.

## Mapping to Code

| Schema Field | Go Struct (`ems-local/internal/sync/aggregator.go`) | TypeScript Type (`ems-cloud/packages/shared/src/`) |
|---|---|---|
| `site_id` | `TelemetryMessage.SiteID` | `TelemetryMessage.site_id` |
| `avg_solar_w` | `TelemetryMessage.AvgSolarW` | `TelemetryMessage.avg_solar_w` |
| ... | (snake_case JSON tags) | (snake_case contract fields) |
