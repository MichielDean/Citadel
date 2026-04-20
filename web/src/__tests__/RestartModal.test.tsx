import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { RestartModal } from '../components/RestartModal';

function mockFetch(response: unknown, ok = true) {
  vi.spyOn(window, 'fetch').mockResolvedValue({
    ok,
    status: ok ? 200 : 500,
    json: () => Promise.resolve(response),
    text: () => Promise.resolve(JSON.stringify(response)),
  } as Response);
}

function getLastFetchBody(): Record<string, unknown> {
  const mock = vi.mocked(window.fetch);
  const calls = mock.mock.calls as [string, RequestInit | undefined][];
  const last = calls[calls.length - 1];
  if (!last[1]?.body) return {};
  return JSON.parse(last[1].body as string);
}

describe('RestartModal', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders nothing when closed', () => {
    const { container } = render(
      <RestartModal open={false} onClose={vi.fn()} dropletId="ct-abc" steps={[]} onRestarted={vi.fn()} />
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('renders step selector when steps provided', () => {
    render(
      <RestartModal open={true} onClose={vi.fn()} dropletId="ct-abc" steps={['flag', 'implement', 'review']} onRestarted={vi.fn()} />
    );
    expect(screen.getByText('Cataractae Step')).toBeInTheDocument();
    expect(screen.getByText('implement')).toBeInTheDocument();
  });

  it('sends cataractae field when step is selected', async () => {
    mockFetch(undefined);

    render(
      <RestartModal open={true} onClose={vi.fn()} dropletId="ct-abc" steps={['flag', 'implement']} onRestarted={vi.fn()} />
    );

    const select = screen.getByRole('combobox');
    await act(() => { fireEvent.change(select, { target: { value: 'implement' } }); });

    const restartBtn = screen.getByRole('button', { name: /restart/i });
    await act(() => { fireEvent.click(restartBtn); });

    expect(window.fetch).toHaveBeenCalled();
    const body = getLastFetchBody();
    expect(body).toEqual({ cataractae: 'implement' });
  });

  it('sends no body when default step is selected', async () => {
    mockFetch(undefined);

    render(
      <RestartModal open={true} onClose={vi.fn()} dropletId="ct-abc" steps={['flag', 'implement']} onRestarted={vi.fn()} />
    );

    const restartBtn = screen.getByRole('button', { name: /restart/i });
    await act(() => { fireEvent.click(restartBtn); });

    expect(window.fetch).toHaveBeenCalled();
    expect(getLastFetchBody()).toEqual({});
  });
});