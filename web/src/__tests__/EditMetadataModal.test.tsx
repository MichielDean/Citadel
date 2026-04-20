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

describe('EditMetadataModal with ComplexitySelector', () => {
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

  it('shows complexity selector with radio options', () => {
    render(
      <EditMetadataModal open={true} onClose={vi.fn()} droplet={baseDroplet} onSaved={vi.fn()} />
    );
    expect(screen.getByText('Standard (1)')).toBeInTheDocument();
    expect(screen.getByText('Full (2)')).toBeInTheDocument();
    expect(screen.getByText('Critical (3)')).toBeInTheDocument();
  });

  it('shows confirmation diff when changes are made', async () => {
    mockFetch({ ...baseDroplet, title: 'Updated Title' });
    const onSaved = vi.fn();
    const onClose = vi.fn();

    render(
      <EditMetadataModal open={true} onClose={onClose} droplet={baseDroplet} onSaved={onSaved} />
    );

    const titleInput = screen.getByDisplayValue('Original Title');
    await act(() => { fireEvent.change(titleInput, { target: { value: 'Updated Title' } }); });

    const saveBtn = screen.getByRole('button', { name: /save/i });
    await act(() => { fireEvent.click(saveBtn); });

    expect(screen.getByText('Confirm Changes')).toBeInTheDocument();
    expect(screen.getByText('Updated Title')).toBeInTheDocument();
  });

  it('calls onClose immediately if no changes were made', async () => {
    const onClose = vi.fn();

    render(
      <EditMetadataModal open={true} onClose={onClose} droplet={baseDroplet} onSaved={vi.fn()} />
    );

    const saveBtn = screen.getByRole('button', { name: /save/i });
    await act(() => { fireEvent.click(saveBtn); });

    expect(onClose).toHaveBeenCalled();
  });
});