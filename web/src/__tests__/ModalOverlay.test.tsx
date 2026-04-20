import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ModalOverlay } from '../components/ModalOverlay';

describe('ModalOverlay', () => {
  it('renders nothing when open is false', () => {
    const { container } = render(
      <ModalOverlay open={false} onClose={vi.fn()}>
        <div>Test Content</div>
      </ModalOverlay>
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('renders children when open is true', () => {
    render(
      <ModalOverlay open={true} onClose={vi.fn()}>
        <div>Test Content</div>
      </ModalOverlay>
    );
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn();
    const { container } = render(
      <ModalOverlay open={true} onClose={onClose}>
        <div>Test Content</div>
      </ModalOverlay>
    );
    const backdrop = container.querySelector('.fixed.inset-0') as HTMLElement;
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });

  it('does not call onClose when content is clicked', () => {
    const onClose = vi.fn();
    render(
      <ModalOverlay open={true} onClose={onClose}>
        <div data-testid="content">Test Content</div>
      </ModalOverlay>
    );
    const content = screen.getByTestId('content');
    content.click();
    expect(onClose).not.toHaveBeenCalled();
  });

  it('has aria-modal attribute', () => {
    const { container } = render(
      <ModalOverlay open={true} onClose={vi.fn()}>
        <div>Test</div>
      </ModalOverlay>
    );
    expect(container.querySelector('[aria-modal="true"]')).toBeInTheDocument();
  });
});