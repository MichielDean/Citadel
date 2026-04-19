import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { DashboardData, Droplet, DropletIssue } from './types';

const API_BASE = '';

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`API ${res.status}: ${text}`);
  }
  return res.json();
}

export function useDashboard() {
  return useQuery<DashboardData>({
    queryKey: ['dashboard'],
    queryFn: () => apiFetch<DashboardData>('/api/dashboard'),
    refetchInterval: 5000,
  });
}

export function useDroplet(id: string) {
  return useQuery<Droplet>({
    queryKey: ['droplet', id],
    queryFn: () => apiFetch<Droplet>(`/api/droplets/${id}`),
    enabled: !!id,
  });
}

export function useDroplets(status?: string) {
  return useQuery<Droplet[]>({
    queryKey: ['droplets', status],
    queryFn: () => {
      const params = status ? `?status=${encodeURIComponent(status)}` : '';
      return apiFetch<Droplet[]>(`/api/droplets${params}`);
    },
  });
}

export function useDropletNotes(id: string) {
  return useQuery({
    queryKey: ['droplet', id, 'notes'],
    queryFn: () => apiFetch(`/api/droplets/${id}/notes`),
    enabled: !!id,
  });
}

export function useDropletIssues(id: string) {
  return useQuery<DropletIssue[]>({
    queryKey: ['droplet', id, 'issues'],
    queryFn: () => apiFetch<DropletIssue[]>(`/api/droplets/${id}/issues`),
    enabled: !!id,
  });
}

export function useSignalDroplet() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, signal, notes }: { id: string; signal: string; notes?: string }) => {
      return apiFetch<Droplet>(`/api/droplets/${id}/${signal}`, {
        method: 'POST',
        body: JSON.stringify(notes ? { notes } : undefined),
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dashboard'] });
      qc.invalidateQueries({ queryKey: ['droplets'] });
    },
  });
}

export function useCastellariusStatus() {
  return useQuery({
    queryKey: ['castellarius', 'status'],
    queryFn: () => apiFetch<{ running: boolean; pid: number; uptime_seconds: number }>('/api/castellarius/status'),
  });
}

export function useCastellariusAction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (action: 'start' | 'stop' | 'restart') => {
      return apiFetch(`/api/castellarius/${action}`, { method: 'POST' });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['castellarius', 'status'] });
    },
  });
}

export function useDoctor() {
  return useQuery({
    queryKey: ['doctor'],
    queryFn: () => apiFetch('/api/doctor'),
    enabled: false,
  });
}

export function useRepos() {
  return useQuery({
    queryKey: ['repos'],
    queryFn: () => apiFetch('/api/repos'),
  });
}

export function useSkills() {
  return useQuery({
    queryKey: ['skills'],
    queryFn: () => apiFetch('/api/skills'),
  });
}