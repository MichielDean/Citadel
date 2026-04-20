import { useState, useEffect } from 'react';
import type { Droplet } from '../api/types';
import { editDroplet } from '../hooks/useApi';
import { ComplexitySelector } from './ComplexitySelector';
import { ModalOverlay } from './ModalOverlay';

interface DiffEntry {
  field: string;
  from: string;
  to: string;
}

interface EditMetadataModalProps {
  open: boolean;
  onClose: () => void;
  droplet: Droplet;
  onSaved: () => void;
}

const COMPLEXITY_LABELS: Record<number, string> = {
  1: 'Standard (1)',
  2: 'Full (2)',
  3: 'Critical (3)',
};

function formatComplexity(v: number): string {
  return COMPLEXITY_LABELS[v] ?? String(v);
}

export function EditMetadataModal({ open, onClose, droplet, onSaved }: EditMetadataModalProps) {
  const [title, setTitle] = useState(droplet.title);
  const [priority, setPriority] = useState(droplet.priority);
  const [complexity, setComplexity] = useState(droplet.complexity);
  const [description, setDescription] = useState(droplet.description);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);

  useEffect(() => {
    if (open) {
      setTitle(droplet.title);
      setPriority(droplet.priority);
      setComplexity(droplet.complexity);
      setDescription(droplet.description);
      setSubmitting(false);
      setError(null);
      setConfirming(false);
    }
  }, [open, droplet.title, droplet.priority, droplet.complexity, droplet.description]);

  const getDiffs = (): DiffEntry[] => {
    const diffs: DiffEntry[] = [];
    if (title !== droplet.title) diffs.push({ field: 'Title', from: droplet.title, to: title });
    if (priority !== droplet.priority) diffs.push({ field: 'Priority', from: String(droplet.priority), to: String(priority) });
    if (complexity !== droplet.complexity) diffs.push({ field: 'Complexity', from: formatComplexity(droplet.complexity), to: formatComplexity(complexity) });
    if (description !== droplet.description) diffs.push({ field: 'Description', from: droplet.description || '(empty)', to: description || '(empty)' });
    return diffs;
  };

  const handleSave = () => {
    const diffs = getDiffs();
    if (diffs.length === 0) {
      onClose();
      return;
    }
    setConfirming(true);
  };

  const handleConfirmSave = async () => {
    setSubmitting(true);
    setError(null);
    try {
      await editDroplet(droplet.id, {
        title: title !== droplet.title ? title : undefined,
        priority: priority !== droplet.priority ? priority : undefined,
        complexity: complexity !== droplet.complexity ? complexity : undefined,
        description: description !== droplet.description ? description : undefined,
      });
      onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update');
    } finally {
      setSubmitting(false);
    }
  };

  if (!open) return null;

  const diffs = getDiffs();

  if (confirming) {
    return (
      <ModalOverlay open={true} onClose={() => setConfirming(false)}>
        <h3 className="font-mono text-cistern-fg text-lg mb-3">Confirm Changes</h3>
        <div className="space-y-2 mb-4">
          {diffs.map((d) => (
            <div key={d.field} className="text-sm">
              <span className="font-mono text-cistern-accent">{d.field}:</span>
              <div className="ml-2 mt-0.5">
                <span className="text-cistern-red line-through">{d.from}</span>
                <span className="mx-1 text-cistern-muted">→</span>
                <span className="text-cistern-green">{d.to}</span>
              </div>
            </div>
          ))}
        </div>
        {error && <div className="mb-3 text-sm text-cistern-red font-mono">{error}</div>}
        <div className="flex gap-2 justify-end">
          <button type="button" onClick={() => setConfirming(false)} className="px-3 py-1.5 text-sm rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg">Back</button>
          <button type="button" onClick={handleConfirmSave} disabled={submitting} className="px-3 py-1.5 text-sm rounded bg-cistern-accent text-cistern-bg font-medium disabled:opacity-50">{submitting ? '…' : 'Confirm Save'}</button>
        </div>
      </ModalOverlay>
    );
  }

  return (
    <ModalOverlay open={true} onClose={onClose}>
      <h3 className="font-mono text-cistern-fg text-lg mb-4">Edit Metadata</h3>

      <div className="space-y-4">
        <div>
          <label className="block text-xs font-mono text-cistern-muted uppercase tracking-wider mb-1">Title</label>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full bg-cistern-bg border border-cistern-border rounded px-2 py-1.5 text-sm text-cistern-fg"
          />
        </div>
        <div>
          <label className="block text-xs font-mono text-cistern-muted uppercase tracking-wider mb-1">Priority</label>
          <input
            type="number"
            value={priority}
            onChange={(e) => setPriority(Number(e.target.value))}
            className="w-full bg-cistern-bg border border-cistern-border rounded px-2 py-1.5 text-sm text-cistern-fg"
          />
        </div>
        <ComplexitySelector value={complexity} onChange={setComplexity} repoName={droplet.repo} />
        <div>
          <label className="block text-xs font-mono text-cistern-muted uppercase tracking-wider mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full bg-cistern-bg border border-cistern-border rounded p-2 text-sm text-cistern-fg resize-y min-h-[80px]"
          />
        </div>
      </div>

      {error && <div className="mt-3 text-sm text-cistern-red font-mono">{error}</div>}

      <div className="flex gap-2 justify-end mt-4">
        <button type="button" onClick={onClose} className="px-3 py-1.5 text-sm rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg">Cancel</button>
        <button type="button" onClick={handleSave} disabled={submitting} className="px-3 py-1.5 text-sm rounded bg-cistern-accent text-cistern-bg font-medium disabled:opacity-50">{submitting ? '…' : 'Save'}</button>
      </div>
    </ModalOverlay>
  );
}