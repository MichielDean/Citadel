import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ReopenModal } from '../components/ReopenModal';

describe('ReopenModal', () => {
  it('renders nothing when closed', () => {
    const { container } = render(
      <ReopenModal open={false} onClose={vi.fn()} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('renders dialog when open', () => {
    render(
      <ReopenModal open={true} onClose={vi.fn()} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    expect(screen.getByText('Reopen Droplet')).toBeInTheDocument();
    expect(screen.getByText(/ct-abc/)).toBeInTheDocument();
  });

  it('calls onConfirm when reopen button is clicked', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(
      <ReopenModal open={true} onClose={vi.fn()} dropletId="ct-abc" onConfirm={onConfirm} />
    );
    const btn = screen.getByRole('button', { name: /reopen/i });
    await fireEvent.click(btn);
    expect(onConfirm).toHaveBeenCalledWith('ct-abc');
  });

  it('calls onClose when cancel is clicked', () => {
    const onClose = vi.fn();
    render(
      <ReopenModal open={true} onClose={onClose} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    screen.getByRole('button', { name: /cancel/i }).click();
    expect(onClose).toHaveBeenCalled();
  });
});