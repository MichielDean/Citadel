import { describe, it, expect } from 'vitest';

function formatAge(iso: string): string {
  const created = new Date(iso).getTime();
  const now = Date.now();
  const diff = now - created;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  return `${days}d`;
}

describe('formatAge', () => {
  it('formats age in minutes for recent items', () => {
    const fiveMinAgo = new Date(Date.now() - 5 * 60000).toISOString();
    const age = formatAge(fiveMinAgo);
    expect(age).toBe('5m');
  });

  it('formats age in hours for older items', () => {
    const twoHoursAgo = new Date(Date.now() - 2 * 3600000).toISOString();
    const age = formatAge(twoHoursAgo);
    expect(age).toBe('2h');
  });

  it('formats age in days for very old items', () => {
    const threeDaysAgo = new Date(Date.now() - 3 * 86400000).toISOString();
    const age = formatAge(threeDaysAgo);
    expect(age).toBe('3d');
  });
});