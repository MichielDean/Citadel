import { useState } from 'react';
import type { DropletDependency } from '../api/types';
import { addDependency, removeDependency } from '../hooks/useApi';

interface DependenciesListProps {
  dropletId: string;
  dependencies: DropletDependency[];
  loading: boolean;
  onChange: () => void;
}

export function DependenciesList({ dropletId, dependencies, loading, onChange }: DependenciesListProps) {
  const [newDepId, setNewDepId] = useState('');
  const [adding, setAdding] = useState(false);
  const [depError, setDepError] = useState<string | null>(null);

  const handleAdd = async () => {
    if (!newDepId.trim()) return;
    setAdding(true);
    setDepError(null);
    try {
      await addDependency(dropletId, newDepId.trim());
      setNewDepId('');
      onChange();
    } catch (err) {
      setDepError(err instanceof Error ? err.message : 'Failed to add dependency');
    } finally {
      setAdding(false);
    }
  };

  const handleRemove = async (depId: string) => {
    setDepError(null);
    try {
      await removeDependency(dropletId, depId);
      onChange();
    } catch (err) {
      setDepError(err instanceof Error ? err.message : 'Failed to remove dependency');
    }
  };

  if (loading) {
    return <div className="text-center py-4 text-cistern-muted font-mono text-sm">Loading dependencies…</div>;
  }

  const blockedBy = dependencies.filter((d) => d.type === 'blocked_by');
  const resolves = dependencies.filter((d) => d.type === 'resolves');
  const blocks = dependencies.filter((d) => d.type === 'blocks');

  return (
    <div className="space-y-4">
      {blockedBy.length > 0 && (
        <div>
          <h4 className="text-xs font-mono text-cistern-muted uppercase tracking-wider mb-2">Blocked By</h4>
          <div className="space-y-1">
            {blockedBy.map((dep) => (
              <div key={dep.depends_on} className="flex items-center gap-2 text-sm">
                <a href={`/app/droplets/${dep.depends_on}`} className="font-mono text-cistern-red hover:underline">{dep.depends_on}</a>
                <button
                  type="button"
                  onClick={() => handleRemove(dep.depends_on)}
                  className="text-cistern-red hover:text-cistern-red/80 text-xs"
                >✕</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {resolves.length > 0 && (
        <div>
          <h4 className="text-xs font-mono text-cistern-muted uppercase tracking-wider mb-2">Resolves</h4>
          <div className="space-y-1">
            {resolves.map((dep) => (
              <div key={dep.depends_on} className="flex items-center gap-2 text-sm">
                <a href={`/app/droplets/${dep.depends_on}`} className="font-mono text-cistern-accent hover:underline">{dep.depends_on}</a>
                <button
                  type="button"
                  onClick={() => handleRemove(dep.depends_on)}
                  className="text-cistern-red hover:text-cistern-red/80 text-xs"
                >✕</button>
              </div>
            ))}
          </div>
        </div>
      )}

      {blocks.length > 0 && (
        <div>
          <h4 className="text-xs font-mono text-cistern-muted uppercase tracking-wider mb-2">Blocks</h4>
          <div className="space-y-1">
            {blocks.map((dep) => (
              <div key={dep.depends_on} className="flex items-center gap-2 text-sm">
                <a href={`/app/droplets/${dep.depends_on}`} className="font-mono text-cistern-accent hover:underline">{dep.depends_on}</a>
              </div>
            ))}
          </div>
        </div>
      )}

      {depError && (
        <div className="text-sm text-cistern-red font-mono">{depError}</div>
      )}

      {dependencies.length === 0 && !depError && (
        <div className="text-center py-2 text-cistern-muted font-mono text-sm">No dependencies</div>
      )}

      <div className="flex gap-2">
        <input
          value={newDepId}
          onChange={(e) => setNewDepId(e.target.value)}
          placeholder="Droplet ID"
          className="flex-1 bg-cistern-bg border border-cistern-border rounded px-2 py-1 text-sm text-cistern-fg font-mono"
          onKeyDown={(e) => { if (e.key === 'Enter') handleAdd(); }}
        />
        <button
          type="button"
          onClick={handleAdd}
          disabled={adding || !newDepId.trim()}
          className="px-3 py-1 text-sm rounded bg-cistern-accent text-cistern-bg font-medium disabled:opacity-50"
        >Add</button>
      </div>
    </div>
  );
}