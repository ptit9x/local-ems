const fs = require('fs');
const path = require('path');
const Ajv = require('ajv/dist/2020');
const addFormats = require('ajv-formats');
const ajv = new Ajv({ allErrors: true }); addFormats(ajv);
const root = path.resolve(__dirname, '../..');
const cases = [
  ['telemetry.schema.json', { schema_version: 1, message_id: 'demo-1', site_id: 'cdi-singapore', timestamp: new Date().toISOString(), interval_sec: 10, cycle_count: 10 }],
  ['commands.schema.json', { schema_version: 1, site_id: 'cdi-singapore', command: 'identify', request_id: 'demo-request', timestamp: new Date().toISOString() }],
  ['heartbeat-v2.schema.json', { schema_version: 2, site_id: 'cdi-singapore', timestamp: new Date().toISOString(), status: 'online', control_owner: 'local', failover_state: 'LOCAL_ACTIVE', lease_epoch: 0, ev_states: [], uptime_seconds: 10 }],
  ['failover-v2.schema.json', { schema_version: 2, site_id: 'cdi-singapore', action: 'takeover', request_id: 'failover-1', timestamp: new Date().toISOString(), lease_epoch: 1 }],
];
for (const [file, value] of cases) {
  const validate = ajv.compile(JSON.parse(fs.readFileSync(path.join(root, 'contracts', file), 'utf8')));
  if (!validate(value)) { console.error(file, validate.errors); process.exitCode = 1; }
}
if (!process.exitCode) console.log('Contract schemas and fixtures are valid.');
