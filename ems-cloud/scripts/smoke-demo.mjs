const base = process.env.DEMO_BASE_URL ?? 'http://localhost:5173/api';

async function json(path, options = {}) {
  const response = await fetch(`${base}${path}`, {
    ...options,
    headers: { 'content-type': 'application/json', ...options.headers },
  });
  if (!response.ok) throw new Error(`${options.method ?? 'GET'} ${path}: HTTP ${response.status} ${await response.text()}`);
  return response.json();
}

const live = await json('/health/live');
if (live.status !== 'ok') throw new Error(`unexpected liveness response: ${JSON.stringify(live)}`);

const login = await json('/auth/login', {
  method: 'POST',
  body: JSON.stringify({ email: 'admin@demo.local', password: 'demo1234' }),
});
if (!login.access_token) throw new Error('login did not return an access token');

const sites = await json('/sites', { headers: { authorization: `Bearer ${login.access_token}` } });
if (!Array.isArray(sites) || sites.length !== 1 || sites[0].siteCode !== 'cdi-singapore') {
  throw new Error(`tenant isolation smoke check failed: ${JSON.stringify(sites)}`);
}

const devices = await json(`/sites/${sites[0].id}/devices`, { headers: { authorization: `Bearer ${login.access_token}` } });
if (!Array.isArray(devices) || devices.length !== 37) {
  throw new Error(`expected 37 seeded devices, received ${devices.length}`);
}

console.log('Demo smoke check passed: health, authentication, tenant isolation, and 37 seeded devices.');
