import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { EditMetadataModal } from '../components/EditMetadataModal';
import type { Droplet } from '../api/types';

const baseDroplet: Droplet = {
  id: 'ct-test1',
  repo: 'cistern',
  title: 'Original Title',
  description: 'Original description',
  priority: 1,
  complexity: 2,
  status: 'in_progress',
  assignee: 'implement',
  current_cataractae: 'implement',
  created_at: '2026-04-19T00:00:00Z',
  updated_at: '2026-04-19T00:00:00Z',
};

function mockFetch(response: unknown, ok = true) {
  vi.spyOn(window, 'fetch').mockResolvedValue({
    ok,
    status: ok ? 200 : 500,
    json: () => Promise.resolve(response),
    text: () => Promise.resolve(JSON.stringify(response)),
  } as Response);
}

describe('EditMetadataModal', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders nothing when closed', () => {
    const { container } = render(
      <EditMetadataModal open={false} onClose={vi.fn()} droplet={baseDroplet} onSaved={vi.fn()} />
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('shows droplet values when opened', () => {
    render(
      <EditMetadataModal open={true} onClose={vi.fn()} droplet={baseDroplet} onSaved={vi.fn()} />
    );
    expect(screen.getByDisplayValue('Original Title')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Original description')).toBeInTheDocument();
  });

  it('resets form values when reopened with changed droplet', () => {
    const updatedDroplet: Droplet = {
      ...baseDroplet,
      title: 'Updated Title',
      description: 'Updated description',
    };

    const { rerender } = render(
      <EditMetadataModal open={true} onClose={vi.fn()} droplet={baseDroplet} onSaved={vi.fn()} />
    );

    expect(screen.getByDisplayValue('Original Title')).toBeInTheDocument();

    rerender(
      <EditMetadataModal open={false} onClose={vi.fn()} droplet={updatedDroplet} onSaved={vi.fn()} />
    );
    expect(screen.queryByDisplayValue('Original Title')).not.toBeInTheDocument();

    rerender(
      <EditMetadataModal open={true} onClose={vi.fn()} droplet={updatedDroplet} onSaved={vi.fn()} />
    );
    expect(screen.getByDisplayValue('Updated Title')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Updated description')).toBeInTheDocument();
  });

  it('calls editDroplet and onSaved on submit', async () => {
    mockFetch(baseDroplet);
    const onSaved = vi.fn();
    const onClose = vi.fn();

    render(
      <EditMetadataModal open={true} onClose={onClose} droplet={baseDroplet} onSaved={onSaved} />
    );

    const titleInput = screen.getByDisplayValue('Original Title');
    await act(() => { fireEvent.change(titleInput, { target: { value: 'New Title' } }); });

    const saveBtn = screen.getByRole('button', { name: /save/i });
    await act(() => { fireEvent.click(saveBtn); });

    expect(window.fetch).toHaveBeenCalled();
    expect(onSaved).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });
});