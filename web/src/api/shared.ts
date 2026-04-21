import { getAuthHeaders, clearStoredKey, isAuthRequired } from '../hooks/useAuth';

export { clearStoredKey } from '../hooks/useAuth';

const MAX_ERROR_BODY_LENGTH = 200;

function truncateErrorMessage(status: number, body: string): string {
  const truncated = body.length > MAX_ERROR_BODY_LENGTH ? body.slice(0, MAX_ERROR_BODY_LENGTH) + '…' : body;
  return `API error ${status}: ${truncated}`;
}

export async function apiFetch<T>(url: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const authHeaders = getAuthHeaders();
  Object.assign(headers, authHeaders);
  const resp = await fetch(url, { ...options, headers: { ...headers, ...options?.headers } });
  if (resp.status === 401) {
    if (isAuthRequired()) {
      clearStoredKey();
      window.dispatchEvent(new CustomEvent('cistern:auth-expired'));
    }
    throw new Error('API error 401: Authentication required');
  }
  if (!resp.ok) {
    const body = await resp.text().catch(() => resp.statusText);
    throw new Error(truncateErrorMessage(resp.status, body));
  }
  if (resp.status === 204) return undefined as T;
  return resp.json();
}