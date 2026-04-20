import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { ComplexitySelector } from '../components/ComplexitySelector';

function mockFetch(urls: Record<string, unknown>) {
  vi.spyOn(window, 'fetch').mockImplementation(async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
    for (const [pattern, response] of Object.entries(urls)) {
      if (url.includes(pattern)) {
        return {
          ok: true,
          status: 200,
          json: () => Promise.resolve(response),
          text: () => Promise.resolve(JSON.stringify(response)),
        } as Response;
      }
    }
    return {
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
      text: () => Promise.resolve('[]'),
    } as Response;
  });
}

describe('ComplexitySelector', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders all three complexity levels', () => {
    render(<ComplexitySelector value={1} onChange={vi.fn()} />);
    expect(screen.getByText('Standard (1)')).toBeInTheDocument();
    expect(screen.getByText('Full (2)')).toBeInTheDocument();
    expect(screen.getByText('Critical (3)')).toBeInTheDocument();
  });

  it('shows the selected value as checked', () => {
    render(<ComplexitySelector value={2} onChange={vi.fn()} />);
    const radios = screen.getAllByRole('radio');
    expect(radios[0]).not.toBeChecked();
    expect(radios[1]).toBeChecked();
    expect(radios[2]).not.toBeChecked();
  });

  it('calls onChange when a level is selected', () => {
    const onChange = vi.fn();
    render(<ComplexitySelector value={1} onChange={onChange} />);
    const criticalRadio = screen.getByRole('radio', { name: 'Critical (3)' });
    fireEvent.click(criticalRadio);
    expect(onChange).toHaveBeenCalledWith(3);
  });

  it('shows fallback pipeline stages when no repoName', () => {
    render(<ComplexitySelector value={1} onChange={vi.fn()} />);
    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('shows all fallback stages for critical complexity', () => {
    render(<ComplexitySelector value={3} onChange={vi.fn()} />);
    expect(screen.getByText('security-review')).toBeInTheDocument();
  });

  it('disables inputs when disabled prop is true', () => {
    render(<ComplexitySelector value={1} onChange={vi.fn()} disabled />);
    const radios = screen.getAllByRole('radio');
    radios.forEach((r) => expect(r).toBeDisabled());
  });

  it('fetches and displays API pipeline stages when repoName provided', async () => {
    const apiSteps = ['implement', 'review', 'qa', 'delivery'];
    mockFetch({ '/api/repos/cistern/steps': apiSteps });

    render(<ComplexitySelector value={1} onChange={vi.fn()} repoName="cistern" />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('shows full API pipeline stages for complexity level 2 when repoName provided', async () => {
    const apiSteps = ['implement', 'review', 'qa', 'delivery'];
    mockFetch({ '/api/repos/myrepo/steps': apiSteps });

    render(<ComplexitySelector value={2} onChange={vi.fn()} repoName="myrepo" />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('review')).toBeInTheDocument();
    expect(screen.getByText('qa')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('shows same API stages for both full and critical when repo provided', async () => {
    const apiSteps = ['implement', 'review', 'qa', 'delivery'];
    mockFetch({ '/api/repos/myrepo/steps': apiSteps });

    render(<ComplexitySelector value={3} onChange={vi.fn()} repoName="myrepo" />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    expect(screen.getByText('qa')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('standard complexity shows first and last API steps when repo provided', async () => {
    const apiSteps = ['implement', 'review', 'qa', 'delivery'];
    mockFetch({ '/api/repos/myrepo/steps': apiSteps });

    render(<ComplexitySelector value={1} onChange={vi.fn()} repoName="myrepo" />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const stageBadges = screen.getAllByText(/^(implement|delivery|review|qa)$/);
    expect(stageBadges).toHaveLength(2);
    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });

  it('falls back to default stages when API fetch fails', async () => {
    vi.spyOn(window, 'fetch').mockRejectedValue(new Error('Network error'));

    render(<ComplexitySelector value={3} onChange={vi.fn()} repoName="badrepo" />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    expect(screen.getByText('implement')).toBeInTheDocument();
    expect(screen.getByText('security-review')).toBeInTheDocument();
    expect(screen.getByText('delivery')).toBeInTheDocument();
  });
});