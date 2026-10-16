/**
 * TelemetryMessage — Aggregated telemetry from Local EMS.
 * Schema: contracts/telemetry.schema.json
 * MQTT Topic: {prefix}/{site_id}/telemetry
 */
export interface TelemetryMessage {
  schema_version: 1;
  message_id: string;
  site_id: string;
  timestamp: string; // ISO 8601 (UTC)
  interval_sec: number;
  avg_solar_w?: number;
  avg_load_w?: number;
  avg_grid_w?: number;
  avg_ess_w?: number;
  min_soc?: number;
  max_soc?: number;
  avg_soc?: number;
  max_temp_c?: number;
  energy_charge_wh?: number;
  energy_discharge_wh?: number;
  cycle_count: number;
  alerts?: string[];
}
