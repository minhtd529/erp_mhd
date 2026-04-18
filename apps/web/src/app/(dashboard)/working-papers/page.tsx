'use client';
import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import { formatDate } from '@/lib/utils';
import type { PaginatedResult } from '@/types';

type WPStatus = 'DRAFT' | 'IN_REVIEW' | 'COMMENTED' | 'FINALIZED' | 'SIGNED_OFF';

interface WorkingPaper {
  id: string;
  title: string;
  engagement_id: string;
  engagement_title?: string;
  status: WPStatus;
  document_type?: string;
  created_by?: string;
  updated_at: string;
}

const STATUS_LABELS: Record<WPStatus, string> = {
  DRAFT: 'Nháp', IN_REVIEW: 'Đang xét duyệt', COMMENTED: 'Có nhận xét',
  FINALIZED: 'Hoàn thiện', SIGNED_OFF: 'Đã ký duyệt',
};
const STATUS_VARIANTS: Record<WPStatus, 'ghost' | 'warning' | 'default' | 'success' | 'secondary'> = {
  DRAFT: 'ghost', IN_REVIEW: 'warning', COMMENTED: 'default',
  FINALIZED: 'success', SIGNED_OFF: 'secondary',
};

export default function WorkingPapersPage() {
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['working-papers', page, q],
    queryFn: () => api.get<PaginatedResult<WorkingPaper>>('/working-papers', { params: { page, size: 20, q: q || undefined } }).then(r => r.data),
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <SearchInput placeholder="Tìm theo tiêu đề..." className="w-80" onSearch={setQ} />
      </div>
      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tiêu đề</TableHead>
                  <TableHead>Hợp đồng</TableHead>
                  <TableHead>Loại tài liệu</TableHead>
                  <TableHead>Người tạo</TableHead>
                  <TableHead>Cập nhật</TableHead>
                  <TableHead>Trạng thái</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((wp) => (
                  <TableRow key={wp.id}>
                    <TableCell className="font-medium">{wp.title}</TableCell>
                    <TableCell className="text-xs text-text-secondary">{wp.engagement_title ?? wp.engagement_id.slice(0, 8)}</TableCell>
                    <TableCell className="text-xs">{wp.document_type ?? '-'}</TableCell>
                    <TableCell className="text-xs">{wp.created_by ?? '-'}</TableCell>
                    <TableCell className="text-xs">{formatDate(wp.updated_at)}</TableCell>
                    <TableCell><Badge variant={STATUS_VARIANTS[wp.status]}>{STATUS_LABELS[wp.status]}</Badge></TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={6} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
      {data && <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />}
    </div>
  );
}
