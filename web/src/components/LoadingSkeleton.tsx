import { type ReactNode } from 'react';

interface LoadingSkeletonWrapperProps {
  loading: boolean;
  children: ReactNode;
  variant?: 'card' | 'row' | 'table';
  count?: number;
  className?: string;
}

export function LoadingSkeleton({ loading, children, variant = 'card', count = 1, className = '' }: LoadingSkeletonWrapperProps) {
  if (!loading) return <>{children}</>;

  if (variant === 'row') {
    return (
      <div className={className}>
        {Array.from({ length: count }).map((_, i) => (
          <div key={i} className="animate-pulse h-6 w-full bg-cistern-border/30 rounded mb-2" />
        ))}
      </div>
    );
  }

  if (variant === 'table') {
    return <SkeletonTable rows={count} cols={4} className={className} />;
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  );
}

export function SkeletonLine({ width = '100%' }: { width?: string }) {
  return <div className="animate-pulse h-4 bg-cistern-border/30 rounded mb-2" style={{ width }} />;
}

export function SkeletonCard({ lines = 3 }: { lines?: number }) {
  return (
    <div className="bg-cistern-surface border border-cistern-border rounded-lg p-4">
      <SkeletonLine width="60%" />
      {Array.from({ length: lines - 1 }).map((_, i) => (
        <SkeletonLine key={i} width={`${90 - i * 20}%`} />
      ))}
    </div>
  );
}

export function SkeletonTable({ rows = 5, cols = 4, className = '' }: { rows?: number; cols?: number; className?: string }) {
  return (
    <div className={`bg-cistern-surface border border-cistern-border rounded-lg overflow-hidden ${className}`}>
      <div className="grid gap-4 p-4" style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}>
        {Array.from({ length: cols }).map((_, i) => (
          <SkeletonLine key={i} width="70%" />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, row) => (
        <div
          key={row}
          className="grid gap-4 p-4 border-t border-cistern-border"
          style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}
        >
          {Array.from({ length: cols }).map((_, col) => (
            <SkeletonLine key={col} width={`${60 + ((row + col) * 7) % 31}%`} />
          ))}
        </div>
      ))}
    </div>
  );
}