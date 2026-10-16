/**
 * RemoteCommand — Command sent from Cloud to Local EMS.
 * Schema: contracts/commands.schema.json
 * MQTT Topic: {prefix}/{site_id}/command
 */
export interface RemoteCommand {
  schema_version: 1;
  site_id: string;
  command: 'update_tariff' | 'update_config' | 'restart' | 'identify';
  request_id: string;
  timestamp: string; // ISO 8601 (UTC)
  payload?: Record<string, unknown>;
}

/**
 * CommandResponse — Response from Local EMS after processing a command.
 * MQTT Topic: {prefix}/{site_id}/command/ack
 */
export interface CommandResponse {
  schema_version: 1;
  request_id: string;
  site_id: string;
  status: 'ok' | 'error';
  timestamp: string;
  result?: Record<string, unknown>;
}
