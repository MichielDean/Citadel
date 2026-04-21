import { useState, useMemo, useCallback } from 'react';
import type { DropletIssue } from '../api/types';
import { IssueCard } from './IssueCard';
import { IssueFilters } from './IssueFilters';
import { ModalOverlay } from './ModalOverlay';
import { SkeletonLine } from './LoadingSkeleton';

interface IssuesListProps {
  issues: DropletIssue[];
  loading: boolean;
  onResolve: (issueId: string, evidence: string) => Promise<void>;
  onReject: (issueId: string, evidence: string) => Promise<void>;
}

export function IssuesList({ issues, loading, onResolve, onReject }: IssuesListProps) {
  const [statusFilter, setStatusFilter] = useState<'all' | 'open'>('open');
  const [roleFilter, setRoleFilter] = useState('');
  const handleRoleFilterChange = useCallback((role: string) => {
    setRoleFilter((prev) => prev === role ? '' : role);
  }, []);
  const [sortOrder, setSortOrder] = useState<'newest' | 'oldest'>('newest');
  const [actionIssue, setActionIssue] = useState<DropletIssue | null>(null);
  const [actionMode, setActionMode] = useState<'resolve' | 'reject'>('resolve');
  const [evidence, setEvidence] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [modalError, setModalError] = useState<string | null>(null);

  const filteredIssues = useMemo(() => {
    let result = issues;

    if (statusFilter === 'open') {
      result = result.filter((i) => i.status === 'open');
    }
    if (roleFilter) {
      result = result.filter((i) => i.flagged_by === roleFilter);
    }

    result = [...result].sort((a, b) => {
      const dateA = new Date(a.flagged_at).getTime();
      const dateB = new Date(b.flagged_at).getTime();
      return sortOrder === 'newest' ? dateB - dateA : dateA - dateB;
    });

    return result;
  }, [issues, statusFilter, roleFilter, sortOrder]);

  if (loading) {
    return <div className="space-y-3 p-4"><SkeletonLine width="100%" /><SkeletonLine width="80%" /><SkeletonLine width="90%" /></div>;
  }

  const openActionModal = (issue: DropletIssue, mode: 'resolve' | 'reject') => {
    setActionIssue(issue);
    setActionMode(mode);
    setEvidence('');
    setModalError(null);
  };

  const handleSubmit = async () => {
    if (!actionIssue || !evidence.trim()) return;
    setSubmitting(true);
    setModalError(null);
    try {
      if (actionMode === 'resolve') {
        await onResolve(actionIssue.id, evidence);
      } else {
        await onReject(actionIssue.id, evidence);
      }
      setActionIssue(null);
      setEvidence('');
    } catch (err) {
      setModalError(err instanceof Error ? err.message : 'Failed to update issue');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <IssueFilters
        issues={issues}
        statusFilter={statusFilter}
        roleFilter={roleFilter}
        sortOrder={sortOrder}
        onStatusFilterChange={setStatusFilter}
        onRoleFilterChange={handleRoleFilterChange}
        onSortOrderChange={setSortOrder}
      />

      {filteredIssues.length === 0 ? (
        <div className="text-center py-4 text-cistern-muted font-mono text-sm">No issues</div>
      ) : (
        <div className="space-y-2">
          {filteredIssues.map((issue) => (
            <IssueCard
              key={issue.id}
              issue={issue}
              onResolve={() => openActionModal(issue, 'resolve')}
              onReject={() => openActionModal(issue, 'reject')}
            />
          ))}
        </div>
      )}

      {actionIssue && (
        <ModalOverlay open={true} onClose={() => setActionIssue(null)}>
          <h3 className="font-mono text-cistern-fg text-lg mb-3">{actionMode === 'resolve' ? 'Resolve' : 'Reject'} Issue</h3>
          <p className="text-sm text-cistern-muted mb-3">{actionIssue.description}</p>
          <textarea
            value={evidence}
            onChange={(e) => setEvidence(e.target.value)}
            placeholder="Evidence (required)..."
            className="w-full bg-cistern-bg border border-cistern-border rounded p-2 text-sm text-cistern-fg resize-y min-h-[80px] mb-3"
            autoFocus
            aria-required="true"
          />
          {modalError && <div className="mb-3 text-sm text-cistern-red font-mono">{modalError}</div>}
          <div className="flex gap-2 justify-end">
            <button type="button" onClick={() => setActionIssue(null)} className="px-3 py-1.5 text-sm rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg">Cancel</button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting || !evidence.trim()}
              className={`px-3 py-1.5 text-sm rounded font-medium disabled:opacity-50 ${
                actionMode === 'resolve'
                  ? 'bg-cistern-green text-cistern-bg'
                  : 'bg-cistern-red text-white'
              }`}
            >{submitting ? '…' : actionMode === 'resolve' ? 'Resolve' : 'Reject'}</button>
          </div>
        </ModalOverlay>
      )}
    </div>
  );
}