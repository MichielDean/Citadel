import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { IssueCard } from '../components/IssueCard';
import type { DropletIssue } from '../api/types';

const openIssue: DropletIssue = {
  id: 'iss-1',
  droplet_id: 'ct-test',
  flagged_by: 'review',
  flagged_at: '2026-04-19T10:00:00Z',
  description: 'Found a bug in the logic',
  status: 'open',
};

const resolvedIssue: DropletIssue = {
  id: 'iss-2',
  droplet_id: 'ct-test',
  flagged_by: 'implement',
  flagged_at: '2026-04-18T10:00:00Z',
  description: 'Another issue\nwith multiline',
  status: 'resolved',
  evidence: 'Fixed in commit abc123',
  resolved_at: '2026-04-19T12:00:00Z',
};

const rejectedIssue: DropletIssue = {
  id: 'iss-3',
  droplet_id: 'ct-test',
  flagged_by: 'qa',
  flagged_at: '2026-04-17T10:00:00Z',
  description: 'Invalid concern',
  status: 'unresolved',
  evidence: 'Not reproducible',
  resolved_at: '2026-04-18T12:00:00Z',
};

describe('IssueCard', () => {
  it('renders issue description', () => {
    render(<IssueCard issue={openIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('Found a bug in the logic')).toBeInTheDocument();
  });

  it('renders status badge with display label', () => {
    render(<IssueCard issue={openIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('Open')).toBeInTheDocument();
  });

  it('renders flagged-by badge', () => {
    render(<IssueCard issue={openIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('review')).toBeInTheDocument();
  });

  it('shows resolve/reject buttons for open issues', () => {
    render(<IssueCard issue={openIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('Resolve')).toBeInTheDocument();
    expect(screen.getByText('Reject')).toBeInTheDocument();
  });

  it('hides resolve/reject buttons for resolved issues', () => {
    render(<IssueCard issue={resolvedIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.queryByText('Resolve')).not.toBeInTheDocument();
    expect(screen.queryByText('Reject')).not.toBeInTheDocument();
  });

  it('shows evidence for resolved issues', () => {
    render(<IssueCard issue={resolvedIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText(/Fixed in commit abc123/)).toBeInTheDocument();
  });

  it('calls onResolve with issue id when resolve is clicked', () => {
    const onResolve = vi.fn();
    render(<IssueCard issue={openIssue} onResolve={onResolve} onReject={vi.fn()} />);
    screen.getByText('Resolve').click();
    expect(onResolve).toHaveBeenCalledWith('iss-1');
  });

  it('calls onReject with issue id when reject is clicked', () => {
    const onReject = vi.fn();
    render(<IssueCard issue={openIssue} onResolve={vi.fn()} onReject={onReject} />);
    screen.getByText('Reject').click();
    expect(onReject).toHaveBeenCalledWith('iss-1');
  });

  it('displays Rejected label for unresolved status', () => {
    render(<IssueCard issue={rejectedIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('Rejected')).toBeInTheDocument();
  });

  it('displays Resolved label for resolved status', () => {
    render(<IssueCard issue={resolvedIssue} onResolve={vi.fn()} onReject={vi.fn()} />);
    expect(screen.getByText('Resolved')).toBeInTheDocument();
  });
});