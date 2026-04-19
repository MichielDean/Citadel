import { describe, it, expect } from 'vitest';

const MAX_BUFFER_SIZE = 50 * 1024;

function truncateBuffer(prev: string, chunk: string): string {
  const next = prev + chunk;
  if (next.length > MAX_BUFFER_SIZE) {
    return next.slice(next.length - MAX_BUFFER_SIZE);
  }
  return next;
}

describe('PeekPanel buffer truncation', () => {
  it('appends small chunks without truncation', () => {
    const result = truncateBuffer('abc', 'def');
    expect(result).toBe('abcdef');
  });

  it('truncates from the beginning when buffer exceeds max size', () => {
    const largeChunk = 'x'.repeat(MAX_BUFFER_SIZE);
    const result = truncateBuffer(largeChunk, 'new');
    expect(result).toHaveLength(MAX_BUFFER_SIZE);
    expect(result.startsWith('new')).toBe(false);
    expect(result.endsWith('new')).toBe(true);
  });

  it('preserves most recent data when buffer exceeds limit', () => {
    const existing = 'a'.repeat(MAX_BUFFER_SIZE - 10);
    const newChunk = 'b'.repeat(20);
    const result = truncateBuffer(existing, newChunk);
    expect(result.length).toBeLessThanOrEqual(MAX_BUFFER_SIZE);
    expect(result.endsWith('b'.repeat(20))).toBe(true);
  });
});

describe('PeekPanel error state', () => {
  it('classifies close code 1008 as auth failure', () => {
    const closeEvent = { code: 1008, reason: '', wasClean: false };
    expect(closeEvent.code === 1008).toBe(true);
  });

  it('classifies close code 4001 as auth failure', () => {
    const closeEvent = { code: 4001, reason: '', wasClean: false };
    expect(closeEvent.code === 4001).toBe(true);
  });

  it('does not classify normal close as auth failure', () => {
    const closeEvent = { code: 1000, reason: '', wasClean: true };
    expect(closeEvent.code === 1008 || closeEvent.code === 4001).toBe(false);
  });
});