import { Component, type ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <div className="flex items-center justify-center h-full p-4">
          <div className="text-center max-w-md">
            <div className="text-cistern-red text-lg font-mono mb-2">Something went wrong</div>
            <div className="text-cistern-muted text-sm mb-4">
              An unexpected error occurred. The error has been logged.
            </div>
            <button
              onClick={() => window.location.reload()}
              className="px-4 py-2 rounded bg-cistern-accent text-cistern-bg text-sm font-medium hover:bg-cistern-accent/90 transition-colors"
            >
              Try Again
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}