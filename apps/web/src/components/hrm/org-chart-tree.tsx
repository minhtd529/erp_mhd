'use client';
import * as React from 'react';
import { cn } from '@/lib/utils';
import type { OrgChartBranch } from '@/services/hrm/organization';

interface OrgChartTreeProps {
  branches: OrgChartBranch[];
}

export function OrgChartTree({ branches }: OrgChartTreeProps) {
  if (branches.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 gap-2 text-text-secondary">
        <span className="text-4xl">🏢</span>
        <p>Tổ chức đang được thiết lập</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-1 p-2">
      {/* Company root node */}
      <div className="flex items-center gap-2 px-3 py-2 rounded bg-primary text-white font-semibold text-sm w-fit mb-2">
        <span>🏢</span>
        <span>MDH Audit Firm</span>
      </div>

      <div className="ml-4 border-l-2 border-border pl-4 flex flex-col gap-3">
        {branches.map(branch => (
          <div key={branch.id}>
            {/* Branch node */}
            <div className={cn(
              'flex items-start gap-2 px-3 py-2 rounded border text-sm w-fit',
              branch.is_head_office
                ? 'border-secondary bg-secondary/10 font-semibold'
                : 'border-border bg-surface'
            )}>
              <span>🏬</span>
              <div>
                <div className="font-medium">{branch.name}</div>
                <div className="text-xs text-text-secondary font-mono">{branch.code}</div>
                {branch.is_head_office && <div className="text-xs text-secondary font-semibold">Trụ sở chính</div>}
              </div>
            </div>

            {/* Dept children */}
            {branch.departments.length > 0 && (
              <div className="ml-4 border-l-2 border-border/50 pl-4 mt-2 flex flex-col gap-1.5">
                {branch.departments.map(dept => (
                  <div key={dept.id} className="flex items-center gap-2 px-3 py-1.5 rounded border border-border bg-background text-xs w-fit">
                    <span>📋</span>
                    <div>
                      <span className="font-medium">{dept.name}</span>
                      <span className="ml-1 text-text-secondary font-mono">({dept.code})</span>
                      <span className={cn(
                        'ml-2 text-[10px] px-1 rounded',
                        dept.dept_type === 'CORE' ? 'bg-action/10 text-action' : 'bg-border text-text-secondary'
                      )}>
                        {dept.dept_type}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
            {branch.departments.length === 0 && (
              <div className="ml-4 border-l-2 border-border/50 pl-4 mt-1">
                <div className="text-xs text-text-secondary italic">Chưa có phòng ban</div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
