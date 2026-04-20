interface IssueFiltersProps {
  issues: { flagged_by: string; status: string }[];
  statusFilter: 'all' | 'open';
  roleFilter: string;
  sortOrder: 'newest' | 'oldest';
  onStatusFilterChange: (filter: 'all' | 'open') => void;
  onRoleFilterChange: (role: string) => void;
  onSortOrderChange: (order: 'newest' | 'oldest') => void;
}

export function IssueFilters({
  issues,
  statusFilter,
  roleFilter,
  sortOrder,
  onStatusFilterChange,
  onRoleFilterChange,
  onSortOrderChange,
}: IssueFiltersProps) {
  const flaggedBySet = new Set(issues.map((i) => i.flagged_by).filter(Boolean));

  return (
    <div className="flex items-center gap-2 mb-3 flex-wrap">
      <FilterBtn active={statusFilter === 'all'} onClick={() => onStatusFilterChange('all')}>All</FilterBtn>
      <FilterBtn active={statusFilter === 'open'} onClick={() => onStatusFilterChange('open')}>Open</FilterBtn>
      {[...flaggedBySet].map((fb) => (
        <FilterBtn key={fb} active={roleFilter === fb} onClick={() => onRoleFilterChange(fb)}>{fb}</FilterBtn>
      ))}
      <div className="ml-auto flex items-center gap-1">
        <label className="text-[10px] font-mono text-cistern-muted uppercase">Sort:</label>
        <select
          value={sortOrder}
          onChange={(e) => onSortOrderChange(e.target.value as 'newest' | 'oldest')}
          className="text-xs bg-cistern-bg border border-cistern-border rounded px-1 py-0.5 text-cistern-fg"
        >
          <option value="newest">Newest</option>
          <option value="oldest">Oldest</option>
        </select>
      </div>
    </div>
  );
}

function FilterBtn({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`text-xs px-2 py-1 rounded-full border transition-colors ${
        active ? 'border-cistern-accent text-cistern-accent bg-cistern-accent/10' : 'border-cistern-border text-cistern-muted hover:text-cistern-fg'
      }`}
    >{children}</button>
  );
}