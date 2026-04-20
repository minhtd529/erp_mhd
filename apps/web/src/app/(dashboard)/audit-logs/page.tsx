'use client';
import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Pagination } from '@/components/shared/pagination';
import { auditService, AUDIT_MODULES, AUDIT_ACTIONS, type AuditLogEntry } from '@/services/audit';
import { formatDateTime } from '@/lib/utils';
import { RotateCcw } from 'lucide-react';

const ACTION_BADGE: Record<string, 'success' | 'danger' | 'warning' | 'default' | 'ghost'> = {
  CREATE: 'success',
  DELETE: 'danger',
  UPDATE: 'default',
  UPDATE_BANK_DETAILS: 'warning',
  UPDATE_STATUS: 'warning',
  ASSIGN_ROLE: 'default',
  APPROVE: 'success',
  REJECT: 'danger',
  SUBMIT: 'default',
  LOCK: 'warning',
  DEACTIVATE: 'ghost',
};

const MODULE_LABELS: Record<string, string> = {
  global: 'Global', org: 'Tổ chức', hrm: 'Nhân sự',
  crm: 'Khách hàng', engagement: 'Hợp đồng', timesheet: 'Chấm công',
  billing: 'Hóa đơn', workingpaper: 'Hồ sơ KT', commission: 'Hoa hồng',
  tax: 'Thuế', reporting: 'Báo cáo',
};

function FilterBar({
  filters,
  onChange,
  onReset,
}: {
  filters: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onReset: () => void;
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Select value={filters.module || 'all'} onValueChange={v => onChange('module', v === 'all' ? '' : v)}>
        <SelectTrigger className="w-40"><SelectValue placeholder="Tất cả module" /></SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Tất cả module</SelectItem>
          {AUDIT_MODULES.map(m => <SelectItem key={m} value={m}>{MODULE_LABELS[m] ?? m}</SelectItem>)}
        </SelectContent>
      </Select>

      <Input
        placeholder="Resource (vd: invoice)"
        value={filters.resource}
        onChange={e => onChange('resource', e.target.value)}
        className="w-44"
      />

      <Select value={filters.action || 'all'} onValueChange={v => onChange('action', v === 'all' ? '' : v)}>
        <SelectTrigger className="w-40"><SelectValue placeholder="Tất cả action" /></SelectTrigger>
        <SelectContent>
          <SelectItem value="all">Tất cả action</SelectItem>
          {AUDIT_ACTIONS.map(a => <SelectItem key={a} value={a}>{a}</SelectItem>)}
        </SelectContent>
      </Select>

      <div className="flex items-center gap-1">
        <Input
          type="date"
          value={filters.from}
          onChange={e => onChange('from', e.target.value)}
          className="w-36"
          title="Từ ngày"
        />
        <span className="text-text-secondary text-sm">–</span>
        <Input
          type="date"
          value={filters.to}
          onChange={e => onChange('to', e.target.value)}
          className="w-36"
          title="Đến ngày"
        />
      </div>

      <Button variant="ghost" size="icon" title="Đặt lại bộ lọc" onClick={onReset}>
        <RotateCcw className="w-4 h-4" />
      </Button>
    </div>
  );
}

const EMPTY_FILTERS = { module: '', resource: '', action: '', from: '', to: '' };

export default function AuditLogsPage() {
  const [page, setPage] = React.useState(1);
  const [filters, setFilters] = React.useState(EMPTY_FILTERS);
  const [applied, setApplied] = React.useState(EMPTY_FILTERS);

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', page, applied],
    queryFn: () => auditService.list({
      page,
      size: 50,
      module: applied.module || undefined,
      resource: applied.resource || undefined,
      action: applied.action || undefined,
      from: applied.from || undefined,
      to: applied.to || undefined,
    }),
  });

  const handleFilterChange = (key: string, value: string) => {
    const next = { ...filters, [key]: value };
    setFilters(next);
    // Apply immediately for selects and date inputs; debounce only for text
    if (key !== 'resource') {
      setApplied(next);
      setPage(1);
    }
  };

  const handleResourceApply = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      setApplied(filters);
      setPage(1);
    }
  };

  const handleReset = () => {
    setFilters(EMPTY_FILTERS);
    setApplied(EMPTY_FILTERS);
    setPage(1);
  };

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">Nhật ký hệ thống</h2>
        <p className="text-sm text-text-secondary">Toàn bộ thao tác thay đổi dữ liệu trong hệ thống (immutable)</p>
      </div>

      <FilterBar
        filters={filters}
        onChange={(k, v) => {
          const next = { ...filters, [k]: v };
          setFilters(next);
          if (k !== 'resource') { setApplied(next); setPage(1); }
        }}
        onReset={handleReset}
      />

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-40">Thời gian</TableHead>
                  <TableHead>Người dùng</TableHead>
                  <TableHead>Module</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>Resource ID</TableHead>
                  <TableHead>Thao tác</TableHead>
                  <TableHead>IP</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((entry) => (
                  <AuditRow key={entry.id} entry={entry} />
                ))}
                {data?.data.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-text-secondary py-10">
                      Không có bản ghi nào
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {data && (
        <div className="flex items-center justify-between">
          <p className="text-xs text-text-secondary">
            {data.total.toLocaleString('vi-VN')} bản ghi
          </p>
          <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />
        </div>
      )}
    </div>
  );
}

function AuditRow({ entry }: { entry: AuditLogEntry }) {
  return (
    <TableRow className="text-xs">
      <TableCell className="text-text-secondary whitespace-nowrap">
        {formatDateTime(entry.created_at)}
      </TableCell>
      <TableCell>
        {entry.user_name
          ? <span className="font-medium">{entry.user_name}</span>
          : <span className="text-text-secondary italic">system</span>
        }
      </TableCell>
      <TableCell>
        <Badge variant="ghost">{MODULE_LABELS[entry.module] ?? entry.module}</Badge>
      </TableCell>
      <TableCell className="font-mono">{entry.resource}</TableCell>
      <TableCell className="font-mono text-text-secondary">
        {entry.resource_id ? entry.resource_id.slice(0, 8) + '…' : '-'}
      </TableCell>
      <TableCell>
        <Badge variant={ACTION_BADGE[entry.action] ?? 'default'}>
          {entry.action}
        </Badge>
      </TableCell>
      <TableCell className="text-text-secondary">{entry.ip_address || '-'}</TableCell>
    </TableRow>
  );
}
