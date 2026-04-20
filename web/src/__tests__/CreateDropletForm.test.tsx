import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { CreateDropletForm } from '../components/CreateDropletForm';
import type { Droplet } from '../api/types';

const mockDroplet: Droplet = {
  id: 'ct-new1',
  repo: 'cistern',
  title: 'New Droplet',
  description: 'A new one',
  priority: 2,
  complexity: 1,
  status: 'open',
  assignee: '',
  current_cataractae: '',
  created_at: '2026-04-19T00:00:00Z',
  updated_at: '2026-04-19T00:00:00Z',
};

const defaultSteps = ['implement', 'review', 'qa', 'docs', 'delivery'];

function mockFetch(urls: Record<string, unknown>) {
  vi.spyOn(window, 'fetch').mockImplementation(async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
    if (url.match(/\/api\/repos\/[^/]+\/steps$/)) {
      const steps = urls['/api/repos/steps'] as string[] | undefined;
      return {
        ok: true, status: 200,
        json: () => Promise.resolve(steps ?? defaultSteps),
        text: () => Promise.resolve(JSON.stringify(steps ?? defaultSteps)),
      } as Response;
    }
    for (const [pattern, response] of Object.entries(urls)) {
      if (pattern === '/api/repos/steps') continue;
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

describe('CreateDropletForm', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => { vi.restoreAllMocks(); });

  it('renders form fields', () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText('Title *')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Priority')).toBeInTheDocument();
    expect(screen.getByText('Complexity')).toBeInTheDocument();
    expect(screen.getByText('Dependencies')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create droplet/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('disables submit when title is empty', () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);
    const submitBtn = screen.getByRole('button', { name: /create droplet/i });
    expect(submitBtn).toBeDisabled();
  });

  it('shows complexity levels', () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText('Standard (1)')).toBeInTheDocument();
    expect(screen.getByText('Full (2)')).toBeInTheDocument();
    expect(screen.getByText('Critical (3)')).toBeInTheDocument();
  });

  it('calls onCancel when cancel is clicked', () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    const onCancel = vi.fn();
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={onCancel} />);
    screen.getByRole('button', { name: /cancel/i }).click();
    expect(onCancel).toHaveBeenCalled();
  });

  it('creates droplet on submit with valid data', async () => {
    mockFetch({
      '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }],
      '/api/droplets': mockDroplet,
    });
    const onSuccess = vi.fn();

    render(<CreateDropletForm onSuccess={onSuccess} onCancel={vi.fn()} />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const textInputs = screen.getAllByRole('textbox');
    const titleInput = textInputs.find((el) => el.getAttribute('required') !== null) ?? textInputs[0];
    await act(() => { fireEvent.change(titleInput, { target: { value: 'New Droplet' } }); });

    const repoSelect = screen.getByRole('combobox');
    await act(() => { fireEvent.change(repoSelect, { target: { value: 'cistern' } }); });

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const submitBtn = screen.getByRole('button', { name: /create droplet/i });
    expect(submitBtn).not.toBeDisabled();
    await act(() => { fireEvent.click(submitBtn); });

    expect(window.fetch).toHaveBeenCalled();
  });

  it('shows error message when creation fails', async () => {
    vi.spyOn(window, 'fetch').mockImplementation(async (input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.match(/\/api\/repos\/[^/]+\/steps$/)) {
        return { ok: true, status: 200, json: () => Promise.resolve(defaultSteps), text: () => Promise.resolve('[]') } as Response;
      }
      if (url.includes('/api/repos')) {
        return {
          ok: true, status: 200,
          json: () => Promise.resolve([{ name: 'cistern', url: 'git@github.com:test/cistern' }]),
          text: () => Promise.resolve('[{"name":"cistern"}]'),
        } as Response;
      }
      if (url.includes('/api/droplets')) {
        return {
          ok: false, status: 500,
          json: () => Promise.resolve({}),
          text: () => Promise.resolve('Internal Server Error'),
        } as Response;
      }
      return { ok: true, status: 200, json: () => Promise.resolve([]), text: () => Promise.resolve('[]') } as Response;
    });
    const onSuccess = vi.fn();

    render(<CreateDropletForm onSuccess={onSuccess} onCancel={vi.fn()} />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const textInputs = screen.getAllByRole('textbox');
    const titleInput = textInputs.find((el) => el.getAttribute('required') !== null) ?? textInputs[0];
    await act(() => { fireEvent.change(titleInput, { target: { value: 'Test Title' } }); });

    const repoSelect = screen.getByRole('combobox');
    await act(() => { fireEvent.change(repoSelect, { target: { value: 'cistern' } }); });

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const submitBtn = screen.getByRole('button', { name: /create droplet/i });
    await act(() => { fireEvent.click(submitBtn); });

    expect(screen.getByText(/API error 500/)).toBeInTheDocument();
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it('shows validation error when title is empty and form is dirty', async () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);

    const titleInput = screen.getAllByRole('textbox').find((el) => el.getAttribute('required') !== null)!;
    fireEvent.change(titleInput, { target: { value: 'A' } });
    fireEvent.change(titleInput, { target: { value: '' } });

    expect(screen.getByText('Title is required')).toBeInTheDocument();
  });

  it('shows validation error when repo is not selected and form is dirty', async () => {
    mockFetch({ '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }] });
    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);

    const titleInput = screen.getAllByRole('textbox').find((el) => el.getAttribute('required') !== null)!;
    await act(() => { fireEvent.change(titleInput, { target: { value: 'Test' } }); });

    expect(screen.getByText('Repo is required')).toBeInTheDocument();
  });

  it('searches for dependencies when typing in dependency search', async () => {
    const searchResults = {
      droplets: [{ id: 'ct-abc', title: 'An existing droplet' }],
      total: 1, page: 1, per_page: 10,
    };
    mockFetch({
      '/api/repos': [{ name: 'cistern', url: 'git@github.com:test/cistern' }],
      '/api/droplets/search': searchResults,
    });

    render(<CreateDropletForm onSuccess={vi.fn()} onCancel={vi.fn()} />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const depInput = screen.getByPlaceholderText('Search droplet ID…');
    await act(() => { fireEvent.change(depInput, { target: { value: 'ct-abc' } }); });

    await act(async () => {
      await new Promise((r) => setTimeout(r, 400));
    });

    expect(screen.getByText('An existing droplet')).toBeInTheDocument();
  });

  it('sends correct payload when creating droplet', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    vi.spyOn(window, 'fetch').mockImplementation(async (input: RequestInfo | URL, options?: RequestInit) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.match(/\/api\/repos\/[^/]+\/steps$/)) {
        return { ok: true, status: 200, json: () => Promise.resolve(defaultSteps), text: () => Promise.resolve('[]') } as Response;
      }
      if (url.includes('/api/repos')) {
        return {
          ok: true, status: 200,
          json: () => Promise.resolve([{ name: 'cistern', url: 'git@github.com:test/cistern' }]),
          text: () => Promise.resolve('[]'),
        } as Response;
      }
      if (url.includes('/api/droplets') && options?.method === 'POST') {
        capturedBody = options?.body ? JSON.parse(String(options.body)) : null;
        return {
          ok: true, status: 200,
          json: () => Promise.resolve(mockDroplet),
          text: () => Promise.resolve(JSON.stringify(mockDroplet)),
        } as Response;
      }
      return { ok: true, status: 200, json: () => Promise.resolve([]), text: () => Promise.resolve('[]') } as Response;
    });

    const onSuccess = vi.fn();
    render(<CreateDropletForm onSuccess={onSuccess} onCancel={vi.fn()} />);

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const titleInput = screen.getAllByRole('textbox').find((el) => el.getAttribute('required') !== null)!;
    await act(() => { fireEvent.change(titleInput, { target: { value: 'New Droplet' } }); });

    const repoSelect = screen.getByRole('combobox');
    await act(() => { fireEvent.change(repoSelect, { target: { value: 'cistern' } }); });

    await act(async () => {
      await new Promise((r) => setTimeout(r, 0));
    });

    const priorityInput = document.querySelector('input[type="number"]')!;
    await act(() => { fireEvent.change(priorityInput, { target: { value: '5' } }); });

    const submitBtn = screen.getByRole('button', { name: /create droplet/i });
    await act(() => { fireEvent.click(submitBtn); });

    expect(capturedBody).toBeTruthy();
    expect(capturedBody!.repo).toBe('cistern');
    expect(capturedBody!.title).toBe('New Droplet');
    expect(capturedBody!.priority).toBe(5);
    expect(capturedBody!.complexity).toBe(1);
  });
});