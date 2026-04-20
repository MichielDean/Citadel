import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { IssuesList } from '../components/IssuesList';
import type { DropletIssue } from '../api/types';

const openIssue: DropletIssue = {
  id: 'issue-1',
  droplet_id: 'ct-abc',
  flagged_by: 'reviewer',
  flagged_at: '2026-04-19T12:00:00Z',
  description: 'Open issue',
  status: 'open',
};

const resolvedIssue: DropletIssue = {
  id: 'issue-2',
  droplet_id: 'ct-abc',
  flagged_by: 'implement',
  flagged_at: '2026-04-18T12:00:00Z',
  description: 'Resolved issue',
  status: 'resolved',
  evidence: 'Fixed in commit abc',
};

const rejectedIssue: DropletIssue = {
  id: 'issue-3',
  droplet_id: 'ct-abc',
  flagged_by: 'qa',
  flagged_at: '2026-04-17T12:00:00Z',
  description: 'Rejected issue',
  status: 'rejected',
  evidence: 'Not applicable',
};

const allIssues = [openIssue, resolvedIssue, rejectedIssue];

describe('IssuesList', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('shows only open issues when status filter is Open', () => {
    render(
      <IssuesList
        issues={allIssues}
        loading={false}
        onResolve={vi.fn()}
        onReject={vi.fn()}
      />
    );

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.queryByText('Resolved issue')).not.toBeInTheDocument();
    expect(screen.queryByText('Rejected issue')).not.toBeInTheDocument();
  });

  it('shows all issues when status filter is switched to All', () => {
    render(
      <IssuesList
        issues={allIssues}
        loading={false}
        onResolve={vi.fn()}
        onReject={vi.fn()}
      />
    );

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.queryByText('Resolved issue')).not.toBeInTheDocument();

    fireEvent.click(screen.getByText('All'));

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.getByText('Resolved issue')).toBeInTheDocument();
    expect(screen.getByText('Rejected issue')).toBeInTheDocument();
  });

  it('toggles role filter off when clicking the same role again', () => {
    render(
      <IssuesList
        issues={allIssues}
        loading={false}
        onResolve={vi.fn()}
        onReject={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('All'));

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.getByText('Resolved issue')).toBeInTheDocument();

    const reviewerBtn = screen.getByRole('button', { name: 'reviewer' });
    fireEvent.click(reviewerBtn);

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.queryByText('Resolved issue')).not.toBeInTheDocument();

    fireEvent.click(reviewerBtn);

    expect(screen.getByText('Open issue')).toBeInTheDocument();
    expect(screen.getByText('Resolved issue')).toBeInTheDocument();
  });
});