const API = import.meta.env.VITE_API_URL ?? '/api';

let accessToken = '';
export function token() { return accessToken; }
export function setToken(t: string) { accessToken = t; }
export function clearToken() { accessToken = ''; }

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

export function request(path: string, options: RequestInit) {
  return fetch(`${API}${path}`, { ...options, credentials: 'include', headers: { 'Content-Type': 'application/json', ...(token() ? { Authorization: `Bearer ${token()}` } : {}), ...options.headers } });
}
