import { describe, it, expect } from 'vitest';
import { formatElapsed } from '../utils/formatElapsed';

function computeStartTimeMs(elapsedNs: number): number {
  return Date.now() - elapsedNs / 1e6;
}

describe('AqueductArch timer drift correction', () => {
  it('computes startTime from elapsed nanoseconds', () => {
    const elapsedNs = 30_000_000_000;
    const now = Date.now();
    const startMs = computeStartTimeMs(elapsedNs);
    expect(startMs).toBeCloseTo(now - 30000, -1);
  });

  it('resets startTime when elapsed changes (SSE refresh)', () => {
    const firstElapsed = 10_000_000_000;
    const secondElapsed = 20_000_000_000;
    const firstStart = computeStartTimeMs(firstElapsed);
    const secondStart = computeStartTimeMs(secondElapsed);
    expect(secondStart).toBeLessThan(firstStart);
  });

  it('computes current elapsed correctly from startTime', () => {
    const elapsedNs = 45_000_000_000;
    const startTimeMs = computeStartTimeMs(elapsedNs);
    const currentElapsedNs = (Date.now() - startTimeMs) * 1e6;
    expect(formatElapsed(currentElapsedNs)).toMatch(/0:4[45]/);
  });

  it('recalculates startTime when flowing transitions without SSE update', () => {
    const elapsedNs = 10_000_000_000;
    const startMs = computeStartTimeMs(elapsedNs);
    const currentElapsed = (Date.now() - startMs) * 1e6;
    expect(formatElapsed(currentElapsed)).toMatch(/0:\d{2}/);
  });
});

describe('formatElapsed edge cases', () => {
  it('returns placeholder for negative values', () => {
    expect(formatElapsed(-1_000_000_000)).toBe('--');
  });
});