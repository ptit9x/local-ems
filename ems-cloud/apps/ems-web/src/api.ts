// Backward-compatible re-exports.
// New code should import from 'lib/api-client' (generic) or 'features/auth/api' (auth).
export { api, token, clearToken } from './lib/api-client';
export { login, restoreSession, logout } from './features/auth/api';
