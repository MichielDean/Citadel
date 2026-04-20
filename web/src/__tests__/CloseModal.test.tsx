import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CloseModal } from '../components/CloseModal';

describe('CloseModal', () => {
  it('renders nothing when closed', () => {
    const { container } = render(
      <CloseModal open={false} onClose={vi.fn()} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('renders dialog when open', () => {
    render(
      <CloseModal open={true} onClose={vi.fn()} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    expect(screen.getByText('Deliver Droplet')).toBeInTheDocument();
    expect(screen.getByText(/ct-abc/)).toBeInTheDocument();
  });

  it('shows warning text', () => {
    render(
      <CloseModal open={true} onClose={vi.fn()} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    expect(screen.getByText(/completed all pipeline stages/)).toBeInTheDocument();
  });

  it('calls onConfirm when deliver button is clicked', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(
      <CloseModal open={true} onClose={vi.fn()} dropletId="ct-abc" onConfirm={onConfirm} />
    );
    const btn = screen.getByRole('button', { name: /deliver/i });
    await fireEvent.click(btn);
    expect(onConfirm).toHaveBeenCalledWith('ct-abc');
  });

  it('calls onClose when cancel is clicked', () => {
    const onClose = vi.fn();
    render(
      <CloseModal open={true} onClose={onClose} dropletId="ct-abc" onConfirm={vi.fn()} />
    );
    screen.getByRole('button', { name: /cancel/i }).click();
    expect(onClose).toHaveBeenCalled();
  });
});