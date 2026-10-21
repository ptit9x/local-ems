# ☁️ EMS Cloud — Multi-Tenant Fleet Management

Cloud platform for managing multiple EMS sites. Receives telemetry from on-premise `ems-local` controllers via MQTT and provides a fleet management dashboard.

## Current Dashboard Scope

The Cloud dashboard is complete as a responsive simulator UI/MVP:

- Vite + React + TypeScript, Tailwind CSS, shared shadcn-style primitives, Radix UI and React Query.
- Login, refresh-session recovery and logout.
- Tenant and site switching, with Tenant B clearly marked as a preview dataset.
- Overview, Live Energy, Sites, Devices, Commands, Alerts, Reports and Settings.
- Profile Settings, Help Center and a responsive notification panel.
- Live API data for sites, telemetry, devices and alerts.

This is not yet a production-complete multi-tenant platform. Site provisioning, database RLS, alert-rule delivery and remote configuration writes are still pending. The verified boundary is documented in [`../docs/plans/implementation-status.md`](../docs/plans/implementation-status.md).

## Quick Start

```bash
# Prerequisites: Node.js 20+, Docker

# 1. Start infrastructure (PostgreSQL/TimescaleDB + Mosquitto)
docker compose up -d postgres mosquitto

# 2. Install dependencies
npm install

# 3. Copy env and configure
cp .env.example .env

# 4. Generate Prisma client & run migrations
npm run db:generate
npm run db:migrate
npm run db:seed -w apps/ems-api

# 5. Start API in dev mode
npm run dev
```

API is available at `http://localhost:3000/api`. Start the web dashboard in another terminal with `npm run dev:web`, then open `http://localhost:5173`.

For the complete demo stack from the repository root, use `make demo`. The seeded admin login is `admin@demo.local` / `demo1234`.

## Architecture

```
                    ┌──────────────────────────┐
                    │      React Dashboard     │
                    │    apps/ems-web/         │
                    └──────────┬───────────────┘
                               │ REST API
                    ┌──────────┴───────────────┐
                    │       NestJS API         │
                    │       apps/ems-api/      │
                    ├──────────────────────────┤
                    │  TelemetryModule         │ ← MQTT subscribe
                    │  SitesModule             │
                    │  HealthModule            │
                    │  AuthModule              │
                    └──────────┬───────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
    ┌─────────┴──┐   ┌────────┴───┐   ┌────────┴───┐
    │ PostgreSQL │   │ TimescaleDB│   │  Mosquitto │
    │  (tenants, │   │ (telemetry │   │   (MQTT    │
    │   users,   │   │  hypertable│   │   broker)  │
    │   sites)   │   │     )      │   │            │
    └────────────┘   └────────────┘   └────────────┘
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health/live` | Process liveness |
| GET | `/api/health/ready` | Database readiness |
| POST | `/api/auth/login` | Sign in and return an access token |
| POST | `/api/auth/refresh` | Refresh access using the HTTP-only cookie |
| POST | `/api/auth/logout` | Revoke the refresh token |
| GET | `/api/sites` | List all sites |
| GET | `/api/sites/:id` | Get site details |
| GET | `/api/sites/:id/devices` | List site devices |
| GET | `/api/sites/:id/alerts` | List site alerts |
| POST | `/api/sites/:id/alerts/:alertId/acknowledge` | Acknowledge an alert |
| GET | `/api/sites/:siteId/telemetry?from=&to=&limit=` | Query telemetry range |
| GET | `/api/sites/:siteId/telemetry/latest` | Latest telemetry record |
| GET | `/api/sites/:siteId/telemetry/daily/:date` | Daily summary |
| POST | `/api/sites/:siteId/commands` | Publish an allowlisted remote command |
| GET | `/api/commands/:id` | Read remote-command status |

All endpoints except health and login/refresh/logout require a bearer access token. Tenant scoping is enforced by the API; Viewer users cannot publish commands.

## MQTT Topics

| Topic | Direction | Description |
|-------|-----------|-------------|
| `ems/{site_id}/telemetry` | Local → Cloud | Aggregated telemetry (interval configurable; demo uses 10 sec) |
| `ems/{site_id}/command` | Cloud → Local | Remote command |
| `ems/{site_id}/command/ack` | Local → Cloud | Command acknowledgement |

Message schemas: [`../contracts/`](../contracts/)

## Project Structure

```
ems-cloud/
├── apps/
│   ├── ems-api/                    # NestJS backend
│   │   ├── src/
│   │   │   ├── main.ts             # Entry point
│   │   │   ├── app.module.ts       # Root module
│   │   │   ├── common/             # Cross-cutting concerns
│   │   │   │   ├── database/       #   Prisma module & service
│   │   │   │   ├── guards/         #   Auth guard
│   │   │   │   ├── pipes/          #   Zod validation pipe
│   │   │   │   └── utils/          #   Shared utilities
│   │   │   └── features/           # Feature modules
│   │   │       ├── auth/           #   JWT auth (controller, service, types)
│   │   │       ├── commands/       #   Remote command lifecycle
│   │   │       ├── failover/       #   Cloud OCPP failover
│   │   │       ├── health/         #   Liveness & readiness probes
│   │   │       ├── sites/          #   Site management
│   │   │       └── telemetry/      #   MQTT ingestion + REST API
│   │   └── prisma/
│   │       └── schema.prisma       # Database schema
│   └── ems-web/                    # React/Vite fleet dashboard
│       └── src/
│           ├── App.tsx             # Root — session restore + routing
│           ├── api.ts              # Backward-compat re-exports
│           ├── components/
│           │   ├── dashboard/      # Shell compositor + shared components
│           │   └── ui/             # shadcn-style primitives
│           ├── features/           # Feature modules
│           │   ├── auth/           #   Login form + auth API
│           │   ├── overview/       #   Fleet overview page
│           │   ├── energy/         #   Live energy page
│           │   ├── sites/          #   Sites management page
│           │   ├── devices/        #   Device inventory page
│           │   ├── commands/       #   Remote commands page
│           │   ├── alerts/         #   Alerts page
│           │   ├── reports/        #   Reports page
│           │   └── settings/       #   Settings, profile, help pages
│           ├── layouts/            # Layout components
│           │   ├── sidebar.tsx     #   Navigation sidebar
│           │   ├── header.tsx      #   Top header bar
│           │   └── notification-menu.tsx  # Notification dropdown
│           └── lib/                # Shared utilities
│               ├── api-client.ts   #   Fetch wrapper + token mgmt
│               └── utils.ts        #   cn() and helpers
├── packages/
│   └── shared/                     # Shared TypeScript types
├── docker-compose.yml              # Dev infrastructure
└── .env.example                    # Environment template
```

## Database Schema

| Table | Description |
|-------|-------------|
| `tenants` | Multi-tenant organizations |
| `users` | Users with roles (ADMIN, OPERATOR, VIEWER) |
| `sites` | EMS sites linked to tenants |
| `telemetry_records` | Time-series telemetry (TimescaleDB hypertable) |
| `devices` | Site device inventory and state |
| `heartbeats` | Site heartbeat history |
| `alerts` | Site alerts and acknowledgements |
| `commands` | Remote-command lifecycle |
| `refresh_tokens` | Revocable authentication sessions |

## Environment Variables

See [`.env.example`](.env.example) for all available configuration.

## Relation to Local EMS

This platform receives data from [`../ems-local/`](../ems-local/) via MQTT. The communication contract is defined in [`../contracts/`](../contracts/). See the root [README](../README.md) for full platform overview.
