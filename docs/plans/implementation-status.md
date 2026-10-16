# Implementation Status

Last verified: 2026-10-16

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
