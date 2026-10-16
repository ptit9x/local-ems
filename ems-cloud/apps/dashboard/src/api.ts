const API = import.meta.env.VITE_API_URL ?? '/api';

let accessToken = '';
export function token() { return accessToken; }
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  let response = await request(path, options);
  if (response.status === 401 && path !== '/auth/login' && path !== '/auth/refresh') {
    const refreshed = await request('/auth/refresh', { method: 'POST' });
    if (refreshed.ok) {
      const result = await refreshed.json() as { access_token: string };
      accessToken = result.access_token;
      response = await request(path, options);
    }
  }
  if (!response.ok) throw new Error((await response.text()) || `HTTP ${response.status}`);
  return response.json();
}
function request(path: string, options: RequestInit) {
  return fetch(`${API}${path}`, { ...options, credentials: 'include', headers: { 'Content-Type': 'application/json', ...(token() ? { Authorization: `Bearer ${token()}` } : {}), ...options.headers } });
}
export async function login(email: string, password: string) {
  const result = await api<{ access_token: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
  accessToken = result.access_token;
}
export async function restoreSession() {
  const response = await request('/auth/refresh', { method: 'POST' });
  if (!response.ok) return false;
  const result = await response.json() as { access_token: string };
  accessToken = result.access_token;
  return true;
}
export async function logout() {
  try { await request('/auth/logout', { method: 'POST' }); } finally { accessToken = ''; }
}
export function clearToken() { accessToken = ''; }
