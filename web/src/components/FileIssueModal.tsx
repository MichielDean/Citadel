import { useState, useEffect } from 'react';
import { addIssue } from '../hooks/useApi';
import { ModalOverlay } from './ModalOverlay';
import type { DropletIssue } from '../api/types';

const FLAGGED_BY_OPTIONS = [
  'implement',
  'review',
  'qa',
  'security-review',
  'docs',
  'delivery',
];

interface FileIssueModalProps {
  open: boolean;
  onClose: () => void;
  dropletId: string;
  onFiled: (issue: DropletIssue) => void;
}

export function FileIssueModal({ open, onClose, dropletId, onFiled }: FileIssueModalProps) {
  const [description, setDescription] = useState('');
  const [flaggedBy, setFlaggedBy] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setDescription('');
      setFlaggedBy('');
      setSubmitting(false);
      setError(null);
    }
  }, [open]);

  const handleSubmit = async () => {
    if (!description.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      const issue = await addIssue(dropletId, {
        description: description.trim(),
        flagged_by: flaggedBy || undefined,
      });
      onFiled(issue);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to file issue');
    } finally {
      setSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && e.ctrlKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <ModalOverlay open={open} onClose={onClose}>
      <h3 className="font-mono text-cistern-fg text-lg mb-4">File Issue</h3>
      <div className="space-y-3" onKeyDown={handleKeyDown}>
        <div>
          <label className="block text-xs font-mono text-cistern-muted uppercase tracking-wider mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe the issue..."
            className="w-full bg-cistern-bg border border-cistern-border rounded p-2 text-sm text-cistern-fg resize-y min-h-[100px]"
            autoFocus
            aria-required="true"
          />
        </div>
        <div>
          <label className="block text-xs font-mono text-cistern-muted uppercase tracking-wider mb-1">Flagged By</label>
          <select
            value={flaggedBy}
            onChange={(e) => setFlaggedBy(e.target.value)}
            className="w-full bg-cistern-bg border border-cistern-border rounded px-2 py-1.5 text-sm text-cistern-fg"
          >
            <option value="">None</option>
            {FLAGGED_BY_OPTIONS.map((opt) => (
              <option key={opt} value={opt}>{opt}</option>
            ))}
          </select>
        </div>
      </div>
      {error && <div className="mt-3 text-sm text-cistern-red font-mono">{error}</div>}
      <div className="flex gap-2 justify-end mt-4">
        <button type="button" onClick={onClose} className="px-3 py-1.5 text-sm rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg">Cancel</button>
        <button type="button" onClick={handleSubmit} disabled={submitting || !description.trim()} className="px-3 py-1.5 text-sm rounded bg-cistern-red text-white font-medium disabled:opacity-50">{submitting ? '…' : 'File Issue'}</button>
      </div>
    </ModalOverlay>
  );
}