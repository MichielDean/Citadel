import { useState, useRef, useEffect } from 'react';

interface RenameInputProps {
  value: string;
  onSave: (newTitle: string) => Promise<void>;
  className?: string;
}

export function RenameInput({ value, onSave, className = '' }: RenameInputProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  const handleSave = async () => {
    const trimmed = draft.trim();
    if (!trimmed || trimmed === value) {
      setDraft(value);
      setEditing(false);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSave(trimmed);
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Rename failed');
    } finally {
      setSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    } else if (e.key === 'Escape') {
      setDraft(value);
      setEditing(false);
    }
  };

  if (!editing) {
    return (
      <h1
        className={`text-xl font-mono font-bold text-cistern-fg truncate cursor-pointer hover:text-cistern-accent transition-colors ${className}`}
        onClick={() => { setDraft(value); setEditing(true); }}
        title="Click to rename"
        role="button"
        tabIndex={0}
        onKeyDown={(e) => { if (e.key === 'Enter') { setDraft(value); setEditing(true); } }}
      >
        {value}
      </h1>
    );
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="text"
        value={draft}
        onChange={(e) => { setDraft(e.target.value); setError(null); }}
        onBlur={handleSave}
        onKeyDown={handleKeyDown}
        disabled={saving}
        className={`text-xl font-mono font-bold text-cistern-fg bg-cistern-bg border rounded px-2 py-1 ${error ? 'border-cistern-red' : 'border-cistern-accent'} ${className}`}
        aria-label="Rename droplet"
        aria-invalid={error ? 'true' : undefined}
      />
      {error && <div className="text-xs text-cistern-red font-mono mt-1">{error}</div>}
    </div>
  );
}