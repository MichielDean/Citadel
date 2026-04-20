import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { FileIssueModal } from '../components/FileIssueModal';
import type { DropletIssue } from '../api/types';

const mockIssue: DropletIssue = {
  id: 'iss-new',
  droplet_id: 'ct-test',
  flagged_by: 'review',
  flagged_at: '2026-04-19T12:00:00Z',
  description: 'Test issue',
  status: 'open',
};

function mockFetch(response: unknown, ok = true) {
  vi.spyOn(window, 'fetch').mockResolvedValue({
    ok,
    status: ok ? 200 : 500,
    json: () => Promise.resolve(response),
    text: () => Promise.resolve(JSON.stringify(response)),
  } as Response);
}

describe('FileIssueModal', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders nothing when closed', () => {
    const { container } = render(
      <FileIssueModal open={false} onClose={vi.fn()} dropletId="ct-test" onFiled={vi.fn()} />
    );
    expect(container.querySelector('.fixed.inset-0')).not.toBeInTheDocument();
  });

  it('renders form when open', () => {
    render(
      <FileIssueModal open={true} onClose={vi.fn()} dropletId="ct-test" onFiled={vi.fn()} />
    );
    expect(screen.getAllByText('File Issue').length).toBeGreaterThan(0);
    expect(screen.getByPlaceholderText('Describe the issue...')).toBeInTheDocument();
    expect(screen.getByText('Flagged By')).toBeInTheDocument();
  });

  it('shows flagged-by dropdown with role options', () => {
    render(
      <FileIssueModal open={true} onClose={vi.fn()} dropletId="ct-test" onFiled={vi.fn()} />
    );
    expect(screen.getByRole('combobox')).toBeInTheDocument();
    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('review')).toBeInTheDocument();
    expect(screen.getByText('qa')).toBeInTheDocument();
    expect(screen.getByText('security-review')).toBeInTheDocument();
    expect(screen.getByText('docs')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('disables submit when description is empty', () => {
    render(
      <FileIssueModal open={true} onClose={vi.fn()} dropletId="ct-test" onFiled={vi.fn()} />
    );
    const btn = screen.getByRole('button', { name: /file issue/i });
    expect(btn).toBeDisabled();
  });

  it('submits issue with description and flagged_by', async () => {
    mockFetch(mockIssue);
    const onFiled = vi.fn();
    const onClose = vi.fn();

    render(
      <FileIssueModal open={true} onClose={onClose} dropletId="ct-test" onFiled={onFiled} />
    );

    const textarea = screen.getByPlaceholderText('Describe the issue...');
    await act(() => { fireEvent.change(textarea, { target: { value: 'Test issue' } }); });

    const select = screen.getByRole('combobox');
    await act(() => { fireEvent.change(select, { target: { value: 'review' } }); });

    const btn = screen.getByRole('button', { name: /file issue/i });
    await act(() => { fireEvent.click(btn); });

    expect(onFiled).toHaveBeenCalledWith(mockIssue);
    expect(onClose).toHaveBeenCalled();
  });
});