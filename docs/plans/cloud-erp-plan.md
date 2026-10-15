# Cloud ERP Development Plan (VPP / Energy Dashboard)

## Goal
Build a centralized Cloud ERP system for real-time monitoring, remote management (Fleet Management), and billing for fleets of Local EMS stations (5MWh BESS) and EV chargers (EV Chargers).

## Technology Stack (Cloud Architecture)

### 1. IoT Ingestion Layer
*   **MQTT Broker:** `EMQX` (Recommended) - Handles millions of connections from Local EMS stations, processes message routing (Rule Engine) to push data directly into the database.

### 2. Database Layer
*   **Time-series DB:** `TimescaleDB` (PostgreSQL extension) - Stores continuously changing telemetry data: Temperature, Grid Power, Battery % (SOC).
*   **Relational DB:** `PostgreSQL` - Stores business data: Customer information, Charging stations, Billing history, Role-based access control (RBAC).
*   **Cache & Real-time:** `Redis` - Caches device status (Online/Offline) and handles Pub/Sub for WebSockets.

### 3. Backend Layer (API & Business Logic)
*   **Framework:** `Node.js` with `NestJS` (TypeScript).
*   **Key Functions:**
    *   Provides REST/GraphQL APIs for the Frontend.
    *   Handles financial logic, reconciliation, and invoicing (Billing).
    *   WebSocket Server receives real-time data and pushes it to the ERP dashboard.
    *   Communicates via standard OCPP Central System (serving fallback/backup connections from EV chargers).

### 4. Frontend Layer (Web ERP / Dashboard)
*   **Core:** `React` + `Vite` (Client-Side Rendering) - Ultra-fast build speed, extremely simple deployment via AWS S3 or Nginx (No server-side rendering required like Next.js).
*   **State Management:** `Zustand` or `Redux Toolkit`.
*   **UI Library:** `Ant Design (AntD)` or `MUI` - Well-suited for building tables and complex configuration forms for ERP.
*   **Charting:** `Apache ECharts` - Renders energy and power flow charts smoothly even with massive amounts of data (tens of thousands of data points).

---

## Main Modules & Tasks (Phase 1 - Cloud MVP)

### Module 1: Cloud & Database Infrastructure (DevOps)
- [ ] Task 1.1: Initialize Docker/Kubernetes cluster. Deploy EMQX broker. → Verify: Test connection successfully using an MQTT Client.
- [ ] Task 1.2: Set up TimescaleDB, PostgreSQL, and Redis database systems. → Verify: Connect successfully to databases using tools (such as DataGrip).

### Module 2: Backend Core (NestJS)
- [ ] Task 2.1: Set up base NestJS project, configure TypeORM/Prisma to connect to PostgreSQL and TimescaleDB. → Verify: `/api/health` API returns 200 OK.
- [ ] Task 2.2: Implement IoT Service, subscribe to messages from EMQX via Webhook or MQTT protocol. → Verify: Logs appear on screen when Local EMS sends simulated data.
- [ ] Task 2.3: Implement logic to insert telemetry data (Current, Voltage, SOC) into TimescaleDB. → Verify: Querying the DB shows data successfully stored in time-series format.

### Module 3: Frontend Web ERP (React + Vite)
- [ ] Task 3.1: Initialize project (`npm create vite@latest`), install Ant Design and React Router. → Verify: Run `npm run dev` to load the default interface.
- [ ] Task 3.2: Build the main ERP Layout (Sidebar, Header, Menu including: General Dashboard, Charging Station Management, Reports). → Verify: UI displays properly according to the layout design.
- [ ] Task 3.3: Build **Multi-site Dashboard** page integrated with ECharts, plotting real-time SOC% charts (receiving data via WebSocket from NestJS). → Verify: Chart updates dynamically and reflects changes immediately upon receiving new data.

### Module 4: Billing & Remote Configuration (Business Features)
- [ ] Task 4.1: Design table schema and APIs to manage electricity tariffs (Time-of-use tariffs). → Verify: Users can create tariff tiers (Peak, Off-peak) through the interface.
- [ ] Task 4.2: Implement electricity price synchronization flow (Remote config) down to Local EMS via MQTT. → Verify: Local EMS logs an alert acknowledging receipt of the new electricity prices.

## Done When
- [ ] The entire Cloud infrastructure is operational (locally).
- [ ] Local EMS (from previous plan) successfully publishes MQTT data to EMQX and persists it to the DB.
- [ ] Management / Operation Engineers log into the Web ERP and can monitor active 5MWh BESS systems with accurate SOC% and Energy Flow charts.
