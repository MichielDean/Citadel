import { describe, it, expect } from 'vitest';
import { LoadingSkeleton, SkeletonLine, SkeletonCard, SkeletonTable } from '../components/LoadingSkeleton';
import { render, screen } from '@testing-library/react';

describe('LoadingSkeleton', () => {
  it('renders children when loading is false', () => {
    render(
      <LoadingSkeleton loading={false}>
        <div>Loaded content</div>
      </LoadingSkeleton>,
    );
    expect(screen.getByText('Loaded content')).toBeInTheDocument();
  });

  it('renders skeleton when loading is true with card variant', () => {
    const { container } = render(
      <LoadingSkeleton loading variant="card">
        <div>Loaded content</div>
      </LoadingSkeleton>,
    );
    expect(container.querySelector('.animate-pulse')).toBeTruthy();
    expect(screen.queryByText('Loaded content')).not.toBeInTheDocument();
  });

  it('renders skeleton when loading is true with row variant', () => {
    const { container } = render(
      <LoadingSkeleton loading variant="row" count={3}>
        <div>Loaded content</div>
      </LoadingSkeleton>,
    );
    const pulses = container.querySelectorAll('.animate-pulse');
    expect(pulses.length).toBe(3);
  });

  it('renders skeleton when loading is true with table variant', () => {
    const { container } = render(
      <LoadingSkeleton loading variant="table" count={2}>
        <div>Loaded content</div>
      </LoadingSkeleton>,
    );
    expect(container.querySelectorAll('.border-t').length).toBe(2);
  });

  it('applies custom className', () => {
    const { container } = render(
      <LoadingSkeleton loading variant="card" className="custom-class">
        <div>Content</div>
      </LoadingSkeleton>,
    );
    expect(container.querySelector('.custom-class')).toBeTruthy();
  });
});

describe('SkeletonLine', () => {
  it('renders a line with default width', () => {
    const { container } = render(<SkeletonLine />);
    const el = container.firstChild as HTMLElement;
    expect(el).toHaveClass('animate-pulse');
    expect(el.style.width).toBe('100%');
  });

  it('renders a line with custom width', () => {
    const { container } = render(<SkeletonLine width="60%" />);
    const el = container.firstChild as HTMLElement;
    expect(el.style.width).toBe('60%');
  });
});

describe('SkeletonCard', () => {
  it('renders a card with default lines', () => {
    const { container } = render(<SkeletonCard />);
    const lines = container.querySelectorAll('.animate-pulse');
    expect(lines.length).toBe(3);
  });

  it('renders a card with specified number of lines', () => {
    const { container } = render(<SkeletonCard lines={5} />);
    const lines = container.querySelectorAll('.animate-pulse');
    expect(lines.length).toBe(5);
  });
});

describe('SkeletonTable', () => {
  it('renders a table with default rows and cols', () => {
    const { container } = render(<SkeletonTable />);
    const rows = container.querySelectorAll('.border-t');
    expect(rows.length).toBe(5);
  });

  it('renders a table with specified rows and cols', () => {
    const { container } = render(<SkeletonTable rows={3} cols={2} />);
    const rows = container.querySelectorAll('.border-t');
    expect(rows.length).toBe(3);
  });
});