export function money(value) {
  return Number(value || 0).toFixed(2);
}

export function signed(value) {
  const number = Number(value || 0);
  return `${number >= 0 ? '+' : ''}${number.toFixed(2)}`;
}

export function formatDate(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

export function signalClass(signal = '') {
  const value = signal.toLowerCase();
  if (value.includes('bull') || value === 'positive' || value === 'complete') return 'positive';
  if (value.includes('bear') || value === 'negative' || value === 'failed') return 'negative';
  if (value.includes('running') || value.includes('info')) return 'info';
  if (value.includes('stale') || value.includes('pending')) return 'warning';
  return 'neutral';
}
