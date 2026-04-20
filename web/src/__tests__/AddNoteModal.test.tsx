import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { AddNoteModal } from '../components/AddNoteModal';

describe('AddNoteModal', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders nothing when closed', () => {
    render(
      <AddNoteModal open={false} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders modal when open', () => {
    render(
      <AddNoteModal open={true} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    expect(screen.getByRole('heading', { name: 'Add Note' })).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Note content...')).toBeInTheDocument();
  });

  it('disables submit when content is empty', () => {
    render(
      <AddNoteModal open={true} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    const submitBtn = screen.getByRole('button', { name: /add note/i });
    expect(submitBtn).toBeDisabled();
  });

  it('enables submit when content is provided', () => {
    render(
      <AddNoteModal open={true} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    fireEvent.change(screen.getByPlaceholderText('Note content...'), { target: { value: 'A note' } });
    const submitBtn = screen.getByRole('button', { name: /add note/i });
    expect(submitBtn).not.toBeDisabled();
  });

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn();
    const { container } = render(
      <AddNoteModal open={true} onClose={onClose} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    const backdrop = container.querySelector('[aria-modal="true"]');
    fireEvent.click(backdrop!);
    expect(onClose).toHaveBeenCalled();
  });

  it('has aria-modal attribute', () => {
    const { container } = render(
      <AddNoteModal open={true} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );
    expect(container.querySelector('[aria-modal="true"]')).toBeInTheDocument();
  });

  it('calls onSaved and onClose on successful submit', async () => {
    vi.spyOn(window, 'fetch').mockResolvedValue({
      ok: true, status: 200,
      json: () => Promise.resolve({}),
      text: () => Promise.resolve('{}'),
    } as Response);
    const onSaved = vi.fn();
    const onClose = vi.fn();

    render(
      <AddNoteModal open={true} onClose={onClose} dropletId="ct-abc" onSaved={onSaved} />
    );

    await act(() => { fireEvent.change(screen.getByPlaceholderText('Note content...'), { target: { value: 'A note' } }); });
    await act(() => { fireEvent.click(screen.getByRole('button', { name: /add note/i })); });

    expect(onSaved).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('shows error when submission fails', async () => {
    vi.spyOn(window, 'fetch').mockResolvedValue({
      ok: false, status: 500,
      json: () => Promise.resolve({}),
      text: () => Promise.resolve('Server error'),
    } as Response);

    render(
      <AddNoteModal open={true} onClose={vi.fn()} dropletId="ct-abc" onSaved={vi.fn()} />
    );

    await act(() => { fireEvent.change(screen.getByPlaceholderText('Note content...'), { target: { value: 'A note' } }); });
    await act(() => { fireEvent.click(screen.getByRole('button', { name: /add note/i })); });

    expect(screen.getByText(/API error 500/)).toBeInTheDocument();
  });
});