const API_BASE = import.meta.env.VITE_API_BASE || '';

export async function api(path, options) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

export function fetchState() {
  return api('/api/state');
}

export function startRun(symbol) {
  return api('/api/runs', {
    method: 'POST',
    body: JSON.stringify({ symbol }),
  });
}

export function fetchRun(id) {
  return api(`/api/runs/${id}`);
}

export function createSnapshot() {
  return api('/api/snapshots', {
    method: 'POST',
    body: JSON.stringify({ note: 'Manual UI snapshot' }),
  });
}
