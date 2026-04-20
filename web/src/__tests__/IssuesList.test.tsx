import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
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
  status: 'unresolved',
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

  describe('resolve evidence modal', () => {
    it('opens evidence modal when resolve button is clicked', () => {
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);

      expect(screen.getByRole('heading', { name: 'Resolve Issue' })).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Evidence (required)...')).toBeInTheDocument();
    });

    it('disables resolve submit when evidence is empty', () => {
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Resolve' && b.closest('[role="dialog"]') && !b.classList.contains('text-xs'))!;
      expect(modalSubmitBtn).toBeDisabled();
    });

    it('enables resolve submit when evidence is provided', () => {
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);

      const textarea = screen.getByPlaceholderText('Evidence (required)...');
      fireEvent.change(textarea, { target: { value: 'Fixed via commit xyz' } });

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Resolve' && b.closest('[role="dialog"]') && !b.classList.contains('text-xs'))!;
      expect(modalSubmitBtn).not.toBeDisabled();
    });

    it('calls onResolve with issue id and evidence', async () => {
      const onResolve = vi.fn().mockResolvedValue(undefined);
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={onResolve}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);

      const textarea = screen.getByPlaceholderText('Evidence (required)...');
      await act(() => { fireEvent.change(textarea, { target: { value: 'Fixed via commit xyz' } }); });

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Resolve' && b.closest('[role="dialog"]') && !b.classList.contains('text-xs'))!;
      await act(() => { fireEvent.click(modalSubmitBtn!); });

      expect(onResolve).toHaveBeenCalledWith('issue-1', 'Fixed via commit xyz');
    });

    it('shows error when onResolve throws', async () => {
      const onResolve = vi.fn().mockRejectedValue(new Error('Server error'));
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={onResolve}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);

      const textarea = screen.getByPlaceholderText('Evidence (required)...');
      await act(() => { fireEvent.change(textarea, { target: { value: 'Evidence' } }); });

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Resolve' && b.closest('[role="dialog"]') && !b.classList.contains('text-xs'))!;
      await act(() => { fireEvent.click(modalSubmitBtn!); });

      expect(screen.getByText('Server error')).toBeInTheDocument();
    });

    it('closes modal on cancel', () => {
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={vi.fn()}
        />
      );

      const resolveBtns = screen.getAllByRole('button', { name: /resolve/i });
      fireEvent.click(resolveBtns[0]);
      expect(screen.getByRole('heading', { name: 'Resolve Issue' })).toBeInTheDocument();

      const dialog = screen.getByRole('dialog');
      const cancelBtn = Array.from(dialog.querySelectorAll('button')).find((b) => b.textContent === 'Cancel')!;
      fireEvent.click(cancelBtn);
      expect(screen.queryByRole('heading', { name: 'Resolve Issue' })).not.toBeInTheDocument();
    });
  });

  describe('reject evidence modal', () => {
    it('opens evidence modal when reject button is clicked', () => {
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={vi.fn()}
        />
      );

      const rejectBtns = screen.getAllByRole('button', { name: /reject/i });
      fireEvent.click(rejectBtns[0]);

      expect(screen.getByRole('heading', { name: 'Reject Issue' })).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Evidence (required)...')).toBeInTheDocument();
    });

    it('calls onReject with issue id and evidence', async () => {
      const onReject = vi.fn().mockResolvedValue(undefined);
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={onReject}
        />
      );

      const rejectBtns = screen.getAllByRole('button', { name: /reject/i });
      fireEvent.click(rejectBtns[0]);

      const textarea = screen.getByPlaceholderText('Evidence (required)...');
      await act(() => { fireEvent.change(textarea, { target: { value: 'Not applicable' } }); });

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Reject' && b.closest('[role="dialog"]'))!;
      await act(() => { fireEvent.click(modalSubmitBtn!); });

      expect(onReject).toHaveBeenCalledWith('issue-1', 'Not applicable');
    });

    it('shows error when onReject throws', async () => {
      const onReject = vi.fn().mockRejectedValue(new Error('Reject failed'));
      render(
        <IssuesList
          issues={[openIssue]}
          loading={false}
          onResolve={vi.fn()}
          onReject={onReject}
        />
      );

      const rejectBtns = screen.getAllByRole('button', { name: /reject/i });
      fireEvent.click(rejectBtns[0]);

      const textarea = screen.getByPlaceholderText('Evidence (required)...');
      await act(() => { fireEvent.change(textarea, { target: { value: 'Evidence' } }); });

      const allButtons = screen.getAllByRole('button');
      const modalSubmitBtn = allButtons.find((b) => b.textContent === 'Reject' && b.closest('[role="dialog"]'))!;
      await act(() => { fireEvent.click(modalSubmitBtn!); });

      expect(screen.getByText('Reject failed')).toBeInTheDocument();
    });
  });
});