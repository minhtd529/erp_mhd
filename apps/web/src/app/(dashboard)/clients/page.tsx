'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import { clientService, type ClientCreateRequest } from '@/services/clients';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Plus, Pencil, Trash2 } from 'lucide-react';

const schema = z.object({
  business_name: z.string().min(1, 'Bắt buộc'),
  tax_code: z.string().min(10, 'Tối thiểu 10 ký tự').max(14, 'Tối đa 14 ký tự'),
  english_name: z.string().optional(),
  address: z.string().min(1, 'Bắt buộc'),
  phone: z.string().optional(),
  email: z.string().email('Email không hợp lệ').optional().or(z.literal('')),
  representative_name: z.string().optional(),
  representative_title: z.string().optional(),
  representative_phone: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

function ClientForm({ initial, onSubmit, loading }: { initial?: Partial<FormData>; onSubmit: (d: FormData) => void; loading?: boolean }) {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: initial,
  });
  return (
    <form id="client-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Tên doanh nghiệp *</Label>
          <Input {...register('business_name')} placeholder="Công ty TNHH..." />
          {errors.business_name && <p className="text-xs text-danger">{errors.business_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Tên tiếng Anh</Label>
          <Input {...register('english_name')} placeholder="Company Ltd." />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Mã số thuế *</Label>
          <Input {...register('tax_code')} placeholder="0123456789" />
          {errors.tax_code && <p className="text-xs text-danger">{errors.tax_code.message}</p>}
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ</Label>
          <Input {...register('address')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email</Label>
          <Input {...register('email')} type="email" />
          {errors.email && <p className="text-xs text-danger">{errors.email.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Người đại diện</Label>
          <Input {...register('representative_name')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chức danh</Label>
          <Input {...register('representative_title')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>SĐT người đại diện</Label>
          <Input {...register('representative_phone')} />
        </div>
      </div>
    </form>
  );
}

export default function ClientsPage() {
  const qc = useQueryClient();
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');
  const [dialog, setDialog] = React.useState<'create' | 'edit' | null>(null);
  const [selected, setSelected] = React.useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['clients', page, q],
    queryFn: () => clientService.list({ page, size: 20, q: q || undefined }),
  });

  const selectedClient = data?.data.find(c => c.id === selected);

  const withOfficeId = (d: FormData): ClientCreateRequest => ({
    ...d,
    office_id: user?.branch_id ?? '',
  });

  const createMutation = useMutation({
    mutationFn: (d: FormData) => clientService.create(withOfficeId(d)),
    onSuccess: () => {
      toast('Tạo khách hàng thành công', 'success');
      qc.invalidateQueries({ queryKey: ['clients'] });
      setDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: (d: FormData) => clientService.update(selected!, withOfficeId(d)),
    onSuccess: () => {
      toast('Cập nhật thành công', 'success');
      qc.invalidateQueries({ queryKey: ['clients'] });
      setDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => clientService.delete(id),
    onSuccess: () => {
      toast('Đã xóa khách hàng', 'success');
      qc.invalidateQueries({ queryKey: ['clients'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <SearchInput placeholder="Tìm theo tên, mã số thuế..." className="w-80" onSearch={setQ} />
        <Button onClick={() => setDialog('create')}><Plus className="w-4 h-4" />Thêm khách hàng</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tên doanh nghiệp</TableHead>
                  <TableHead>Mã số thuế</TableHead>
                  <TableHead>Người đại diện</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Ngày tạo</TableHead>
                  <TableHead className="w-20"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((client) => (
                  <TableRow key={client.id}>
                    <TableCell>
                      <p className="font-medium">{client.business_name}</p>
                      {client.english_name && <p className="text-xs text-text-secondary">{client.english_name}</p>}
                    </TableCell>
                    <TableCell className="font-mono text-xs">{client.tax_code}</TableCell>
                    <TableCell>{client.representative_name ?? '-'}</TableCell>
                    <TableCell className="text-xs">{client.email ?? '-'}</TableCell>
                    <TableCell className="text-xs">{formatDate(client.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" onClick={() => { setSelected(client.id); setDialog('edit'); }}>
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          variant="ghost" size="icon"
                          className="text-danger hover:text-danger"
                          onClick={() => { if (confirm('Xóa khách hàng này?')) deleteMutation.mutate(client.id); }}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </TableCell>
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

      <Dialog open={dialog === 'create'} onOpenChange={(o) => !o && setDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Thêm khách hàng mới</DialogTitle></DialogHeader>
          <ClientForm onSubmit={(d) => createMutation.mutate(d)} loading={createMutation.isPending} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="client-form" loading={createMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={dialog === 'edit'} onOpenChange={(o) => !o && setDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Chỉnh sửa khách hàng</DialogTitle></DialogHeader>
          {selectedClient && (
            <ClientForm
              initial={selectedClient}
              onSubmit={(d) => updateMutation.mutate(d)}
              loading={updateMutation.isPending}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="client-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
