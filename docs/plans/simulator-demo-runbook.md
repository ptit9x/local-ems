# Simulator Demo Runbook

This is the active implementation runbook. Older planning documents are historical references.

## Start

Requirements: Docker with Compose, Go 1.27+ and Node.js 20+.

```bash
make demo
```

Open `http://localhost:5173` and sign in with `admin@demo.local` / `demo1234`.
The Local dashboard remains available at `http://localhost:8080`.

Wait for the Cloud API to finish applying the schema and seed data, then run:

```bash
make demo-check
```

This verifies API health, authentication, tenant isolation, and the 37 seeded devices through the same Nginx route used by the browser.

The demo models 4 BMS, 24 stub Lite-On PCS units, 6 EV chargers and 3 meters. It does not use or claim compatibility with physical register maps.

## Verify

```bash
make test-all
curl http://localhost:3000/api/health/live
curl http://localhost:3000/api/health/ready
```

Telemetry is aggregated every 10 seconds. Stop Mosquitto to exercise the local SQLite outbox, then restart it and confirm pending records drain without duplicate Cloud records.

## Stop or reset

`make demo-down` stops Cloud containers and preserves demo data. `make demo-reset` deletes Docker volumes and the Local SQLite demo database; this cannot be undone.

## Known limitations

- All device data is simulated with deterministic topology but time-varying load data.
- Lite-On registers remain an explicit stub pending vendor documentation.
- Cloud dashboard refreshes every five seconds; WebSocket streaming is deferred.
- Physical network integration, device calibration and site acceptance are out of scope.
