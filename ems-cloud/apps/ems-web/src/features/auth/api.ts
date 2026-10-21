import { api, request, setToken, clearToken } from "../../lib/api-client";

export async function login(email: string, password: string) {
  const result = await api<{ access_token: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
  setToken(result.access_token);
}

export async function restoreSession() {
  const response = await request('/auth/refresh', { method: 'POST' });
  if (!response.ok) return false;
  const result = await response.json() as { access_token: string };
  setToken(result.access_token);
  return true;
}

export async function logout() {
  try { await request('/auth/logout', { method: 'POST' }); } finally { clearToken(); }
}

export { api } from "../../lib/api-client";
