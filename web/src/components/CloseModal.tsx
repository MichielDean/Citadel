import { useState, useEffect } from 'react';
import { ModalOverlay } from './ModalOverlay';

interface CloseModalProps {
  open: boolean;
  onClose: () => void;
  dropletId: string;
  onConfirm: (dropletId: string) => Promise<void>;
}

export function CloseModal({ open, onClose, dropletId, onConfirm }: CloseModalProps) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setSubmitting(false);
      setError(null);
    }
  }, [open]);

  const handleConfirm = async () => {
    setSubmitting(true);
    setError(null);
    try {
      await onConfirm(dropletId);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to close droplet');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ModalOverlay open={open} onClose={onClose}>
      <h3 className="font-mono text-cistern-fg text-lg mb-2">Deliver Droplet</h3>
      <p className="text-sm text-cistern-muted mb-2">
        This will mark the droplet <span className="font-mono text-cistern-accent">{dropletId}</span> as delivered.
      </p>
      <p className="text-sm text-cistern-yellow mb-4">
        This action should only be used when the droplet has completed all pipeline stages.
      </p>
      {error && <div className="mb-3 text-sm text-cistern-red font-mono">{error}</div>}
      <div className="flex gap-2 justify-end">
        <button type="button" onClick={onClose} className="px-3 py-1.5 text-sm rounded border border-cistern-border text-cistern-muted hover:text-cistern-fg">Cancel</button>
        <button type="button" onClick={handleConfirm} disabled={submitting} className="px-3 py-1.5 text-sm rounded bg-cistern-accent text-cistern-bg font-medium disabled:opacity-50">{submitting ? '…' : 'Deliver'}</button>
      </div>
    </ModalOverlay>
  );
}