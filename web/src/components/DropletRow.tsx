import type { Droplet } from '../api/types';
import { StatusBadge } from './StatusBadge';

interface DropletRowProps {
  droplet: Droplet;
  blockedBy?: string;
}

export function DropletRow({ droplet, blockedBy }: DropletRowProps) {
  const age = formatAge(droplet.created_at);

  return (
    <div
      className="w-full flex items-center gap-3 px-3 py-2 rounded-md hover:bg-cistern-border/20 transition-colors text-left"
    >
      <StatusBadge status={droplet.status} />
      <span className="font-mono text-xs text-cistern-accent">{droplet.id}</span>
      <span className="text-sm text-cistern-fg truncate flex-1">{droplet.title}</span>
      {blockedBy && (
        <span className="text-xs text-cistern-yellow" title={`Blocked by ${blockedBy}`}>
          ⛏ {blockedBy}
        </span>
      )}
      <span className="text-xs text-cistern-muted whitespace-nowrap">{age}</span>
    </div>
  );
}

function formatAge(iso: string): string {
  const created = new Date(iso).getTime();
  const now = Date.now();
  const diff = now - created;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  return `${days}d`;
}