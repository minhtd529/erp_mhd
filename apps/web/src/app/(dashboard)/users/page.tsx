'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import { userService, type UserCreateRequest, type UserUpdateRequest } from '@/services/users';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { Plus, Pencil, Trash2, UserCog } from 'lucide-react';

const ROLES = [
  { code: 'SUPER_ADMIN', label: 'Super Admin' },
  { code: 'FIRM_PARTNER', label: 'Partner' },
  { code: 'AUDIT_MANAGER', label: 'Audit Manager' },
  { code: 'AUDIT_STAFF', label: 'Audit Staff' },
  { code: 'CLIENT_ADMIN', label: 'Client Admin' },
  { code: 'CLIENT_USER', label: 'Client User' },
];

const STATUS_OPTIONS = [
  { value: 'active', label: 'Hoạt động' },
  { value: 'inactive', label: 'Không hoạt động' },
  { value: 'locked', label: 'Bị khoá' },
];

const STATUS_BADGE: Record<string, 'success' | 'ghost' | 'warning' | 'danger'> = {
  active: 'success',
  inactive: 'ghost',
  locked: 'danger',
};

const createSchema = z.object({
  full_name: z.string().min(1, 'Bắt buộc'),
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(12, 'Tối thiểu 12 ký tự'),
  role_code: z.string().min(1, 'Chọn vai trò'),
});

const editSchema = z.object({
  full_name: z.string().min(1, 'Bắt buộc'),
  status: z.enum(['active', 'inactive', 'locked']),
});

type CreateFormData = z.infer<typeof createSchema>;
type EditFormData = z.infer<typeof editSchema>;

function CreateUserForm({ onSubmit }: { onSubmit: (d: CreateFormData) => void }) {
  const { register, handleSubmit, control, formState: { errors } } = useForm<CreateFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: { role_code: 'AUDIT_STAFF' },
  });
  return (
    <form id="create-user-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Họ và tên *</Label>
          <Input {...register('full_name')} placeholder="Nguyễn Văn A" />
          {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email *</Label>
          <Input {...register('email')} type="email" placeholder="user@company.com" />
          {errors.email && <p className="text-xs text-danger">{errors.email.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Mật khẩu *</Label>
          <Input {...register('password')} type="password" placeholder="Tối thiểu 12 ký tự" />
          {errors.password && <p className="text-xs text-danger">{errors.password.message}</p>}
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Vai trò ban đầu *</Label>
          <Controller
            control={control}
            name="role_code"
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger><SelectValue placeholder="Chọn vai trò" /></SelectTrigger>
                <SelectContent>
                  {ROLES.map(r => <SelectItem key={r.code} value={r.code}>{r.label}</SelectItem>)}
                </SelectContent>
              </Select>
            )}
          />
          {errors.role_code && <p className="text-xs text-danger">{errors.role_code.message}</p>}
        </div>
      </div>
    </form>
  );
}

function EditUserForm({ initial, onSubmit }: { initial: EditFormData; onSubmit: (d: EditFormData) => void }) {
  const { register, handleSubmit, control, formState: { errors } } = useForm<EditFormData>({
    resolver: zodResolver(editSchema),
    defaultValues: initial,
  });
  return (
    <form id="edit-user-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <Label>Họ và tên *</Label>
        <Input {...register('full_name')} />
        {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
      </div>
      <div className="flex flex-col gap-1">
        <Label>Trạng thái *</Label>
        <Controller
          control={control}
          name="status"
          render={({ field }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {STATUS_OPTIONS.map(s => <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>)}
              </SelectContent>
            </Select>
          )}
        />
        {errors.status && <p className="text-xs text-danger">{errors.status.message}</p>}
      </div>
    </form>
  );
}

export default function UsersPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');
  const [statusFilter, setStatusFilter] = React.useState('');
  const [dialog, setDialog] = React.useState<'create' | 'edit' | 'role' | null>(null);
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const [assignRoleCode, setAssignRoleCode] = React.useState('AUDIT_STAFF');

  const { data, isLoading } = useQuery({
    queryKey: ['users', page, q, statusFilter],
    queryFn: () => userService.list({ page, size: 20, q: q || undefined, status: statusFilter || undefined }),
  });

  const selectedUser = data?.data.find(u => u.id === selectedId);

  const createMut = useMutation({
    mutationFn: (d: UserCreateRequest) => userService.create(d),
    onSuccess: () => {
      toast('Tạo người dùng thành công', 'success');
      qc.invalidateQueries({ queryKey: ['users'] });
      setDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMut = useMutation({
    mutationFn: (d: UserUpdateRequest) => userService.update(selectedId!, d),
    onSuccess: () => {
      toast('Cập nhật thành công', 'success');
      qc.invalidateQueries({ queryKey: ['users'] });
      setDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => userService.delete(id),
    onSuccess: () => {
      toast('Đã xoá người dùng', 'success');
      qc.invalidateQueries({ queryKey: ['users'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const assignRoleMut = useMutation({
    mutationFn: () => userService.assignRole(selectedId!, assignRoleCode),
    onSuccess: () => {
      toast('Đã gán vai trò thành công', 'success');
      qc.invalidateQueries({ queryKey: ['users'] });
      setDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <SearchInput
            placeholder="Tìm theo tên, email..."
            className="w-72"
            onSearch={v => { setQ(v); setPage(1); }}
          />
          <Select
            value={statusFilter || 'all'}
            onValueChange={v => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}
          >
            <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tất cả trạng thái</SelectItem>
              {STATUS_OPTIONS.map(s => <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
        <Button onClick={() => setDialog('create')}>
          <Plus className="w-4 h-4" />Thêm người dùng
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Họ tên</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead>2FA</TableHead>
                  <TableHead className="w-28"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.full_name}</TableCell>
                    <TableCell className="text-xs">{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={STATUS_BADGE[user.status] ?? 'ghost'}>
                        {STATUS_OPTIONS.find(s => s.value === user.status)?.label ?? user.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={user.two_factor_enabled ? 'success' : 'ghost'}>
                        {user.two_factor_enabled ? 'Bật' : 'Tắt'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost" size="icon" title="Chỉnh sửa"
                          onClick={() => { setSelectedId(user.id); setDialog('edit'); }}
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          variant="ghost" size="icon" title="Gán vai trò"
                          onClick={() => { setSelectedId(user.id); setAssignRoleCode('AUDIT_STAFF'); setDialog('role'); }}
                        >
                          <UserCog className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          variant="ghost" size="icon" title="Xoá"
                          className="text-danger hover:text-danger"
                          onClick={() => {
                            if (confirm(`Xoá người dùng "${user.full_name}"?`)) deleteMut.mutate(user.id);
                          }}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-text-secondary py-8">
                      Không có dữ liệu
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {data && <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />}

      {/* Create Dialog */}
      <Dialog open={dialog === 'create'} onOpenChange={o => !o && setDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Thêm người dùng mới</DialogTitle></DialogHeader>
          <CreateUserForm onSubmit={d => createMut.mutate(d)} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="create-user-form" loading={createMut.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={dialog === 'edit'} onOpenChange={o => !o && setDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Chỉnh sửa người dùng</DialogTitle></DialogHeader>
          {selectedUser && (
            <EditUserForm
              initial={{ full_name: selectedUser.full_name, status: selectedUser.status }}
              onSubmit={d => updateMut.mutate(d)}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="edit-user-form" loading={updateMut.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Assign Role Dialog */}
      <Dialog open={dialog === 'role'} onOpenChange={o => !o && setDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Gán vai trò — {selectedUser?.full_name}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-2">
            <Label>Vai trò</Label>
            <Select value={assignRoleCode} onValueChange={setAssignRoleCode}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {ROLES.map(r => <SelectItem key={r.code} value={r.code}>{r.label}</SelectItem>)}
              </SelectContent>
            </Select>
            <p className="text-xs text-text-secondary">
              Hành động này sẽ thêm vai trò vào người dùng (không xoá vai trò hiện tại).
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button onClick={() => assignRoleMut.mutate()} loading={assignRoleMut.isPending}>
              Gán vai trò
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
