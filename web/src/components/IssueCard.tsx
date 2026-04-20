import type { DropletIssue } from '../api/types';
import { formatAge } from '../utils/formatAge';

const FLAGGED_BY_COLORS: Record<string, string> = {
  implement: 'bg-cistern-accent/20 text-cistern-accent',
  review: 'bg-cistern-yellow/20 text-cistern-yellow',
  qa: 'bg-cistern-green/20 text-cistern-green',
  'security-review': 'bg-cistern-red/20 text-cistern-red',
  docs: 'bg-blue-500/20 text-blue-400',
  delivery: 'bg-purple-500/20 text-purple-400',
};

const STATUS_DISPLAY: Record<string, { label: string; color: string }> = {
  open: { label: 'Open', color: 'bg-cistern-yellow/20 text-cistern-yellow' },
  resolved: { label: 'Resolved', color: 'bg-cistern-green/20 text-cistern-green' },
  unresolved: { label: 'Rejected', color: 'bg-cistern-red/20 text-cistern-red' },
};

interface IssueCardProps {
  issue: DropletIssue;
  onResolve: (issueId: string) => void;
  onReject: (issueId: string) => void;
}

export function IssueCard({ issue, onResolve, onReject }: IssueCardProps) {
  const statusInfo = STATUS_DISPLAY[issue.status] ?? { label: issue.status, color: 'bg-cistern-muted/20 text-cistern-muted' };
  const flaggedByColor = FLAGGED_BY_COLORS[issue.flagged_by] ?? 'bg-cistern-muted/20 text-cistern-muted';

  return (
    <div className="bg-cistern-surface border border-cistern-border rounded-lg p-3">
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <span className={`text-xs px-1.5 py-0.5 rounded-full font-mono font-medium ${statusInfo.color}`}>
          {statusInfo.label}
        </span>
        <span className={`text-xs px-1.5 py-0.5 rounded-full font-mono ${flaggedByColor}`}>
          {issue.flagged_by}
        </span>
        <span className="text-[10px] font-mono text-cistern-muted ml-auto">
          {formatAge(issue.flagged_at)}
        </span>
      </div>
      <div className="text-sm text-cistern-fg whitespace-pre-wrap mb-2">{issue.description}</div>
      {issue.evidence && (
        <div className="text-xs text-cistern-muted bg-cistern-bg/50 rounded p-2 mb-2 whitespace-pre-wrap">
          <span className="font-mono uppercase tracking-wider text-[10px] mr-1">Evidence:</span>
          {issue.evidence}
        </div>
      )}
      {issue.resolved_at && (
        <div className="text-[10px] text-cistern-muted font-mono">
          Resolved {formatAge(issue.resolved_at)}
        </div>
      )}
      {issue.status === 'open' && (
        <div className="flex gap-2 mt-2">
          <button
            type="button"
            onClick={() => onResolve(issue.id)}
            className="text-xs px-2 py-1 rounded border border-cistern-green/40 text-cistern-green hover:bg-cistern-green/10 transition-colors"
          >Resolve</button>
          <button
            type="button"
            onClick={() => onReject(issue.id)}
            className="text-xs px-2 py-1 rounded border border-cistern-red/40 text-cistern-red hover:bg-cistern-red/10 transition-colors"
          >Reject</button>
        </div>
      )}
    </div>
  );
}