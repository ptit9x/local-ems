# Architecture overview image prompt

`architecture_simple.png` is the stakeholder-friendly, icon-rich Phase 2 overview.
It is intentionally a designed raster infographic rather than a Mermaid render.

## Regeneration prompt

Create a polished 16:9 dark-navy EMS architecture infographic with three vertical
zones: **SITE HARDWARE — SINGAPORE**, **LOCAL EMS — AUTONOMOUS EDGE**, and
**EMS CLOUD**.

- Site: recognizable hardware illustrations for Power Meter ×3, BMS ×4,
  PCS ×24 — Lite-On, EV Charger ×6, and an Industrial Ethernet Switch.
- Local: a visually simple "Expandable Adapter Layer" overview label, Data
  Acquisition, P1–P4 Constraint Control. Keep implementation detail such as
  AdapterFactory/DeviceRegistry in the technical Mermaid diagram, not this overview.
  Engine, Command Distributor, Local OCPP Central System, Failover Agent with
  `lease_epoch`, SQLite 7-day Buffer, MQTT Store-and-Forward, Hot-Reload Config,
  and Local Dashboard.
- Cloud: MQTT TLS Broker, NestJS API, PostgreSQL/TimescaleDB, Cloud OCPP + EV
  Failover Controller, FDD + Alerts, Email/SMS/Webhook, and React Dashboard.
- Show solid primary OCPP from EV to Local, cyan MQTT telemetry/heartbeat/commands,
  takeover/handover, and a dashed coral **Backup OCPP 1.6J — Phase 2** path from
  EV directly to Cloud.
- Include the notes **Logical backup path — vendor capability gate** and
  **BESS control continues during Cloud outage**.
- Use cyan edge modules, green cloud modules, yellow control priorities, purple
  databases, and coral backup connections. No watermark or decorative fake text.

The detailed deterministic diagram remains `mermaid.mmd` → `architecture.png`.

