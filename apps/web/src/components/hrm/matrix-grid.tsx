'use client';
import * as React from 'react';
import { cn } from '@/lib/utils';
import type { Branch, Department, BranchDepartment } from '@/services/hrm/organization';

interface MatrixGridProps {
  branches: Branch[];
  departments: Department[];
  links: BranchDepartment[];
  canWrite: boolean;
  onLink: (branchId: string, deptId: string) => void;
  onUnlink: (branchId: string, deptId: string) => void;
}

export function MatrixGrid({ branches, departments, links, canWrite, onLink, onUnlink }: MatrixGridProps) {
  const linked = React.useMemo(() => {
    const set = new Set<string>();
    links.forEach(l => { if (l.is_active) set.add(`${l.branch_id}:${l.department_id}`); });
    return set;
  }, [links]);

  if (branches.length === 0 || departments.length === 0) {
    return (
      <div className="flex items-center justify-center py-16 text-text-secondary">
        Chưa có dữ liệu để hiển thị ma trận.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="text-xs border-collapse">
        <thead>
          <tr>
            <th className="bg-surface border border-border px-3 py-2 text-left min-w-32 text-text-primary font-semibold">
              Chi nhánh \ Phòng ban
            </th>
            {departments.map(d => (
              <th key={d.id} className="bg-surface border border-border px-2 py-2 text-center min-w-20 font-medium text-text-primary">
                <div className="font-mono text-[10px] text-text-secondary">{d.code}</div>
                <div className="text-xs truncate max-w-[80px]">{d.name}</div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {branches.map((b, bi) => (
            <tr key={b.id} className={bi % 2 === 0 ? 'bg-white' : 'bg-background/50'}>
              <td className="border border-border px-3 py-2 font-medium">
                <div className="font-mono text-[10px] text-text-secondary">{b.code}</div>
                <div>{b.name}</div>
                {b.is_head_office && <div className="text-[10px] text-secondary font-semibold">Trụ sở</div>}
              </td>
              {departments.map(d => {
                const key = `${b.id}:${d.id}`;
                const isLinked = linked.has(key);
                return (
                  <td key={d.id} className="border border-border text-center p-0">
                    <button
                      disabled={!canWrite}
                      onClick={() => isLinked ? onUnlink(b.id, d.id) : onLink(b.id, d.id)}
                      className={cn(
                        'w-full h-10 flex items-center justify-center transition-colors',
                        isLinked
                          ? 'text-success bg-success/10 hover:bg-success/20'
                          : 'text-text-secondary hover:bg-background',
                        !canWrite && 'cursor-default'
                      )}
                      title={isLinked ? 'Nhấp để hủy liên kết' : 'Nhấp để liên kết'}
                    >
                      {isLinked ? '✓' : '–'}
                    </button>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
