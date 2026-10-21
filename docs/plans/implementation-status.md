# Implementation Status

Last verified: 2026-10-21

This is the source of truth for implementation status. Older plans and proposals describe the intended roadmap and may retain their original estimates.

| Area | Implemented | Automated verification | Remaining before production |
|---|---|---|---|
| Local control engine | Constraint resolver, safety/grid/optimization controllers, one-second cycle, PCS distribution | Go unit tests and `go vet` | Physical-device commissioning and calibrated site settings |
| Simulator | 4 BMS, 24 stub Lite-On PCS, 6 EV chargers, 3 meters, scripted scenario | Simulator and scenario unit tests | Load/soak test with production-equivalent gateway hardware |
| Local persistence | SQLite WAL, retention, disk monitor, telemetry outbox and replay | Store, retention and sync tests | Operational backup/recovery procedure |
| Local ↔ Cloud messaging | MQTT QoS 1 telemetry, heartbeat, allowlisted commands and acknowledgements | Go sync tests and JSON Schema validation | Production broker authentication, certificate provisioning and rotation |
| Cloud API | JWT/refresh auth, RBAC, tenant-scoped sites/devices/alerts/telemetry/commands | Jest service tests for auth, tenant isolation, ingestion and commands | Broader controller/database integration tests and performance tests |
| Cloud ingestion | Telemetry, heartbeat, acknowledgement handling, offline detection, database deduplication | Unit tests plus database uniqueness constraints | Long-running reconnect and broker failover test |
| Cloud dashboard | Premium responsive React/Vite UI; login/session refresh; tenant and site switchers; overview, live energy, sites, devices, commands, alerts, reports, settings, profile, help center and notification views | TypeScript production build, React Doctor 100/100, desktop/mobile browser QA and six 1440×900 documentation screenshots | Wire demo-only actions to production APIs, add route-level code splitting, complete accessibility/browser matrix and add UI E2E tests |
| Demo packaging | Docker Compose, TimescaleDB, Mosquitto, API, Nginx dashboard and seed data | Container build, `make demo-check`, live Local → MQTT → Cloud telemetry, and command acknowledgement verified | Repeat the smoke test in CI or before each tagged demo release |
| CI | Go tests/race/vet, Cloud build/tests and contract validation | GitHub Actions workflow | Optional containerized end-to-end job |

## Phase 2: EV integration and Cloud failover

### Current progress

| Workstream | Status | Verified result | Remaining exit criteria |
|---|---|---|---|
| OCPP protocol core | ✅ Software complete | Local Central System is wired into physical runtime; Boot/Heartbeat/Status/Meter/Start/Stop and command responses are covered; Cloud backup endpoint has six-charger integration coverage | TLS termination, production credentials and real charger/HIL verification |
| EV Dynamic Load Management | ✅ Software complete | Fair-share limits are dispatched through live EV adapters with lease gating, bounded retry, duplicate suppression and six simulated OCPP charger feedback/command E2E | Verify electrical behavior and limits against real chargers during HIL/SAT |
| Failover contracts/state | ✅ Software core | Heartbeat/failover v2, explicit owner, monotonic lease epoch, persisted sessions/events and stale-command rejection | Exercise the complete workflow against six connected charger simulators |
| Cloud takeover/handover | ✅ Software core | 90-second detector, safe-mode command, monotonic lease, startup recovery guard, stale-epoch Cloud reassertion and handover persistence are implemented and tested | Full process-kill Docker E2E, uninterrupted-session HIL and SAT |
| Time-of-Use | 🟡 Partial | Local ToU controller and allowlisted remote tariff update are implemented and tested | Site timezone, weekday/holiday overrides, overlap validation and audited revision UI |
| Runtime configuration | 🟡 Partial | EV site limit/enabled values validate and apply atomically; unknown keys are rejected | Device IP/port/protocol and topology diff/apply, health check, persisted revision and rollback |
| AdapterFactory + DeviceRegistry | ✅ Runtime integrated | Selects by device type/protocol/vendor; builds transactionally; runtime polls normalized snapshots into the aggregator and dispatches PCS/EV commands | Physical-device commissioning and long-duration reconnect/soak testing |
| FDD and notifications | 🔴 Not complete | Offline alert record exists; Prisma `AlertRule` schema is prepared | Rule CRUD/evaluator, debounce/cooldown, delivery workers for email/SMS/webhook, retry/audit and dashboard UI |
| Operations dashboard | 🟡 MVP | Existing fleet/site/device/command/alert views build successfully | EV sessions, control owner, failover timeline, alert rules and config deployment views/API wiring |
| Architecture/documentation | ✅ Complete for target design | Detailed Mermaid/PNG plus icon-rich stakeholder PNG and regeneration prompt | Keep status annotations aligned with implementation; vendor topology is still an external input |
| E2E, HIL and SAT | 🔴 Not started | Unit/race/vet/build/contract checks pass | Six-charger Docker E2E, soak/reconnect tests, vendor gate, HIL and Singapore SAT |

Implemented in the software baseline:

- Versioned heartbeat/failover contracts with explicit control owner and monotonic `lease_epoch`.
- Local failover state machine and MQTT control subscription; stale lease commands are rejected.
- Cloud offline detection creates persisted takeover sessions/events and publishes safe-mode takeover commands.
- Recovery heartbeat/handover returns control to Local and records the transition timeline.
- EV DLM fair-share allocator enforces site headroom and excludes offline chargers.
- Opt-in Cloud OCPP 1.6J WebSocket endpoint maintains a charger registry and applies
  safe-mode `SetChargingProfile` commands when Cloud receives the control lease.
- Remote EV site-limit configuration is allowlisted, validated and applied atomically.
- Config-driven `AdapterFactory` selects by device type/protocol/vendor, creates
  Meter/BMS Modbus, PCS profile-backed and EV Modbus/OCPP branches, then builds
  `DeviceRegistry` transactionally with normalized health/snapshots.
- Registry runtime drives physical polling, normalized aggregation and lease-gated
  PCS/EV commands with bounded retry and duplicate EV-limit suppression.
- Six simulated OCPP chargers complete Boot, Status, StartTransaction, MeterValues,
  `SetChargingProfile` and StopTransaction in an automated race-tested E2E.
- Restarted Local enters `HANDOVER_PENDING`; stale recovery cannot steal an active
  Cloud lease, which prevents command emission before authoritative handback.
- Prisma models/migration for failover sessions/events, EV sessions, config revisions and alert rules.
- Full Phase 2 technical architecture generated reproducibly from Mermaid, plus an
  icon-rich stakeholder overview with a versioned regeneration prompt.

Remaining before Phase 2 is software-ready:

- Harden the Cloud OCPP endpoint with integration tests, TLS termination, per-charger
  credentials and the vendor-confirmed physical endpoint-switch strategy.
- Run the complete Local-process kill/restart + MQTT + Cloud + six-charger workflow
  in Docker and retain timing evidence for the 90-second takeover/30-second handback SLOs.
- Complete atomic device/topology hot reload with persisted revision, health check and rollback.
- Alert-rule evaluation and email/SMS/webhook delivery workers plus dashboard management views.
- Add dashboard views for EV sessions, control ownership, failover timeline and config deployments.
- Docker E2E for six OCPP chargers, forced takeover, split-brain prevention, handback and reconnect soak.
- Vendor capability gate, HIL and Singapore SAT.

### Phase 2 acceptance status

- [ ] Six chargers connect to Local OCPP and appear on the Cloud dashboard.
- [ ] Killing Local EMS causes Cloud safe-mode takeover within 90 seconds.
- [ ] Lease enforcement proves Local and Cloud cannot both issue valid commands.
- [ ] Restarting Local returns control within 30 seconds without interrupting an active session.
- [ ] Device/topology config reload succeeds without restart and rolls back on failed health check.
- [ ] Fault/offline/fire rules produce one deduplicated alert and audited notification delivery.
- [ ] Vendor capability gate, HIL and Singapore SAT are signed off.

## Phase 9: Multi-tenant Cloud status

Phase 9C (Cloud Dashboard) is complete at the simulator UI/MVP level. This does **not** mean that the entire Phase 9 platform is production-complete.

### UI/MVP implemented

- Vite + React + TypeScript dashboard using Tailwind CSS, shared shadcn-style primitives, Radix UI and React Query.
- Authentication flow from login to a protected dashboard, including refresh-session recovery and logout.
- Tenant switcher with live Demo Tenant A and a clearly labelled preview dataset for Demo Tenant B.
- Responsive application shell and functional navigation for Overview, Live Energy, Sites, Devices, Commands, Alerts, Reports and Settings.
- Profile Settings, Help Center and responsive notification panel with unread state, mark-as-read and link to Alerts.
- Live Cloud API integration for sites, telemetry, devices and alerts; polling and cache invalidation are handled through React Query.
- Desktop and mobile visual QA. Cloud screenshots are stored in `docs/screenshots/cloud-*.png`.

### Remaining before Phase 9 can be called production-complete

- Real user registration/invitation and tenant membership administration.
- Complete site CRUD/provisioning, including per-site MQTT credentials and generated Local EMS configuration.
- Database-enforced row-level security; current tenant isolation is enforced by authenticated application queries.
- Alert-rule CRUD, rule evaluation configuration and email/webhook delivery.
- Remote topology/control-strategy configuration with audited MQTT delivery to Local EMS.
- Route-level URL navigation, UI end-to-end tests, production accessibility/browser matrix and frontend bundle splitting.

## External blockers

- The Lite-On register map is a deliberate stub until the vendor supplies the exact model documentation.
- Physical Modbus/OCPP addresses, network topology and credentials must come from the Singapore site/vendor.
- Calibration, fire-system interlocks and site acceptance require physical equipment and authorized operating procedures.

Until these blockers are resolved, the deliverable is a simulator MVP and must not be represented as a commissioned production EMS.
