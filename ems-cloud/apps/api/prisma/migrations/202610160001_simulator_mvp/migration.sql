CREATE EXTENSION IF NOT EXISTS timescaledb;
-- Prisma creates the relational tables from schema.prisma. This idempotent block
-- promotes time-series tables after `prisma db push`/migration has created them.
DO $$ BEGIN
  IF to_regclass('public.telemetry_records') IS NOT NULL THEN
    PERFORM create_hypertable('telemetry_records', 'timestamp', if_not_exists => TRUE, migrate_data => TRUE);
  END IF;
  IF to_regclass('public.heartbeats') IS NOT NULL THEN
    PERFORM create_hypertable('heartbeats', 'timestamp', if_not_exists => TRUE, migrate_data => TRUE);
  END IF;
END $$;
