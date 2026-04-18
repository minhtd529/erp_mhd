'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import { engagementService, type EngagementStatus } from '@/services/engagements';
import { clientService } from '@/services/clients';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate, formatCurrency } from '@/lib/utils';
import { Plus, ChevronRight } from 'lucide-react';

const STATUS_LABELS: Record<EngagementStatus, string> = {
  DRAFT: 'Nháp', PROPOSAL: 'Đề xuất', CONTRACTED: 'Đã ký HĐ',
  ACTIVE: 'Đang thực hiện', COMPLETED: 'Hoàn thành', SETTLED: 'Đã quyết toán',
};
const STATUS_VARIANTS: Record<EngagementStatus, 'ghost' | 'warning' | 'default' | 'success' | 'outline' | 'secondary'> = {
  DRAFT: 'ghost', PROPOSAL: 'warning', CONTRACTED: 'default',
  ACTIVE: 'success', COMPLETED: 'secondary', SETTLED: 'outline',
};
const ACTION_LABELS: Record<string, string> = {
  submit: 'Gửi đề xuất', contract: 'Ký hợp đồng', activate: 'Bắt đầu',
  complete: 'Hoàn thành', settle: 'Quyết toán', reject: 'Từ chối',
};

const schema = z.object({
  client_id: z.string().min(1, 'Chọn khách hàng'),
  title: z.string().min(1, 'Bắt buộc'),
  description: z.string().optional(),
  service_type: z.string().optional(),
  start_date: z.string().optional(),
  end_date: z.string().optional(),
  budget: z.coerce.number().optional(),
});
type FormData = z.infer<typeof schema>;

export default function EngagementsPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');
  const [dialog, setDialog] = React.useState(false);
  const [transDialog, setTransDialog] = React.useState<{ id: string; status: EngagementStatus } | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['engagements', page, q],
    queryFn: () => engagementService.list({ page, size: 20, q: q || undefined }),
  });
  const { data: clients } = useQuery({
    queryKey: ['clients-all'],
    queryFn: () => clientService.list({ size: 100 }),
  });

  const { register, handleSubmit, formState: { errors }, reset } = useForm<FormData>({ resolver: zodResolver(schema) });

  const createMutation = useMutation({
    mutationFn: (d: FormData) => engagementService.create(d),
    onSuccess: () => { toast('Tạo hợp đồng thành công', 'success'); qc.invalidateQueries({ queryKey: ['engagements'] }); setDialog(false); reset(); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const transitionMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: string }) => engagementService.transition(id, action),
    onSuccess: () => { toast('Cập nhật trạng thái thành công', 'success'); qc.invalidateQueries({ queryKey: ['engagements'] }); setTransDialog(null); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <SearchInput placeholder="Tìm theo tiêu đề, mô tả..." className="w-80" onSearch={setQ} />
        <Button onClick={() => setDialog(true)}><Plus className="w-4 h-4" />Tạo hợp đồng</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Hợp đồng</TableHead>
                  <TableHead>Khách hàng</TableHead>
                  <TableHead>Loại dịch vụ</TableHead>
                  <TableHead>Ngân sách</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="w-24"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((eng) => {
                  const transitions = engagementService.getAvailableTransitions(eng.status);
                  return (
                    <TableRow key={eng.id}>
                      <TableCell>
                        <p className="font-medium">{eng.title}</p>
                        <p className="text-xs text-text-secondary font-mono">{eng.code}</p>
                      </TableCell>
                      <TableCell className="text-sm">{eng.client_name ?? '-'}</TableCell>
                      <TableCell className="text-xs">{eng.service_type ?? '-'}</TableCell>
                      <TableCell className="text-xs">{eng.budget ? formatCurrency(eng.budget) : '-'}</TableCell>
                      <TableCell>
                        <Badge variant={STATUS_VARIANTS[eng.status]}>{STATUS_LABELS[eng.status]}</Badge>
                      </TableCell>
                      <TableCell>
                        {transitions.length > 0 && (
                          <Button
                            variant="ghost" size="sm"
                            onClick={() => setTransDialog({ id: eng.id, status: eng.status })}
                          >
                            <ChevronRight className="w-3.5 h-3.5" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={6} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
      {data && <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />}

      <Dialog open={dialog} onOpenChange={setDialog}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Tạo hợp đồng mới</DialogTitle></DialogHeader>
          <form id="eng-form" onSubmit={handleSubmit(d => createMutation.mutate(d))} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Khách hàng *</Label>
              <select {...register('client_id')} className="flex h-9 w-full rounded border border-border bg-surface px-3 text-sm focus:outline-none focus:border-action">
                <option value="">Chọn khách hàng</option>
                {clients?.data.map(c => <option key={c.id} value={c.id}>{c.business_name}</option>)}
              </select>
              {errors.client_id && <p className="text-xs text-danger">{errors.client_id.message}</p>}
            </div>
            <div className="flex flex-col gap-1">
              <Label>Tiêu đề *</Label>
              <Input {...register('title')} />
              {errors.title && <p className="text-xs text-danger">{errors.title.message}</p>}
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1">
                <Label>Loại dịch vụ</Label>
                <Input {...register('service_type')} placeholder="Kiểm toán BCTC..." />
              </div>
              <div className="flex flex-col gap-1">
                <Label>Ngân sách (VND)</Label>
                <Input {...register('budget')} type="number" />
              </div>
              <div className="flex flex-col gap-1">
                <Label>Ngày bắt đầu</Label>
                <Input {...register('start_date')} type="date" />
              </div>
              <div className="flex flex-col gap-1">
                <Label>Ngày kết thúc</Label>
                <Input {...register('end_date')} type="date" />
              </div>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Mô tả</Label>
              <Textarea {...register('description')} rows={3} />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(false)}>Hủy</Button>
            <Button type="submit" form="eng-form" loading={createMutation.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!transDialog} onOpenChange={(o) => !o && setTransDialog(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Chuyển trạng thái</DialogTitle></DialogHeader>
          <div className="flex flex-col gap-2">
            {transDialog && engagementService.getAvailableTransitions(transDialog.status).map(action => (
              <Button
                key={action}
                variant={action === 'reject' ? 'danger' : 'primary'}
                className="w-full"
                loading={transitionMutation.isPending}
                onClick={() => transitionMutation.mutate({ id: transDialog.id, action })}
              >
                {ACTION_LABELS[action] ?? action}
              </Button>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
