'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import {
  branchService, departmentService,
  type BranchCreateRequest, type BranchUpdateRequest,
  type DepartmentCreateRequest, type DepartmentUpdateRequest,
  type Branch,
} from '@/services/branches';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { Plus, Pencil } from 'lucide-react';

// ─── Schemas ──────────────────────────────────────────────────────────────────

const branchCreateSchema = z.object({
  code: z.string().min(1, 'Bắt buộc').max(20, 'Tối đa 20 ký tự'),
  name: z.string().min(1, 'Bắt buộc').max(200),
  address: z.string().optional(),
  phone: z.string().optional(),
});

const branchUpdateSchema = branchCreateSchema.extend({
  is_active: z.boolean(),
});

const deptCreateSchema = z.object({
  code: z.string().min(1, 'Bắt buộc').max(20, 'Tối đa 20 ký tự'),
  name: z.string().min(1, 'Bắt buộc').max(200),
  branch_id: z.string().optional(),
});

const deptUpdateSchema = deptCreateSchema.extend({
  is_active: z.boolean(),
});

type BranchCreateForm = z.infer<typeof branchCreateSchema>;
type BranchUpdateForm = z.infer<typeof branchUpdateSchema>;
type DeptCreateForm = z.infer<typeof deptCreateSchema>;
type DeptUpdateForm = z.infer<typeof deptUpdateSchema>;

// ─── Branch forms ─────────────────────────────────────────────────────────────

function BranchCreateForm({ onSubmit }: { onSubmit: (d: BranchCreateForm) => void }) {
  const { register, handleSubmit, formState: { errors } } = useForm<BranchCreateForm>({
    resolver: zodResolver(branchCreateSchema),
  });
  return (
    <form id="branch-create-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label>Mã chi nhánh *</Label>
          <Input {...register('code')} placeholder="HN01" />
          {errors.code && <p className="text-xs text-danger">{errors.code.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} placeholder="024 xxxx xxxx" />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Tên chi nhánh *</Label>
          <Input {...register('name')} placeholder="Chi nhánh Hà Nội" />
          {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ</Label>
          <Input {...register('address')} placeholder="123 Đường ABC, Quận XYZ" />
        </div>
      </div>
    </form>
  );
}

function BranchEditForm({ initial, onSubmit }: { initial: BranchUpdateForm; onSubmit: (d: BranchUpdateForm) => void }) {
  const { register, handleSubmit, control, formState: { errors } } = useForm<BranchUpdateForm>({
    resolver: zodResolver(branchUpdateSchema),
    defaultValues: initial,
  });
  return (
    <form id="branch-edit-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label>Mã chi nhánh *</Label>
          <Input {...register('code')} />
          {errors.code && <p className="text-xs text-danger">{errors.code.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Tên chi nhánh *</Label>
          <Input {...register('name')} />
          {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ</Label>
          <Input {...register('address')} />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Trạng thái</Label>
          <Controller
            control={control}
            name="is_active"
            render={({ field }) => (
              <Select
                value={field.value ? 'true' : 'false'}
                onValueChange={v => field.onChange(v === 'true')}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">Đang hoạt động</SelectItem>
                  <SelectItem value="false">Ngừng hoạt động</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>
      </div>
    </form>
  );
}

// ─── Department forms ─────────────────────────────────────────────────────────

function DeptCreateForm({ branches, onSubmit }: { branches: Branch[]; onSubmit: (d: DeptCreateForm) => void }) {
  const { register, handleSubmit, control, formState: { errors } } = useForm<DeptCreateForm>({
    resolver: zodResolver(deptCreateSchema),
  });
  return (
    <form id="dept-create-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label>Mã phòng ban *</Label>
          <Input {...register('code')} placeholder="KT01" />
          {errors.code && <p className="text-xs text-danger">{errors.code.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chi nhánh</Label>
          <Controller
            control={control}
            name="branch_id"
            render={({ field }) => (
              <Select value={field.value ?? 'none'} onValueChange={v => field.onChange(v === 'none' ? undefined : v)}>
                <SelectTrigger><SelectValue placeholder="Chọn chi nhánh" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Không thuộc chi nhánh</SelectItem>
                  {branches.map(b => <SelectItem key={b.id} value={b.id}>{b.name}</SelectItem>)}
                </SelectContent>
              </Select>
            )}
          />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Tên phòng ban *</Label>
          <Input {...register('name')} placeholder="Phòng Kế toán" />
          {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
        </div>
      </div>
    </form>
  );
}

function DeptEditForm({ branches, initial, onSubmit }: { branches: Branch[]; initial: DeptUpdateForm; onSubmit: (d: DeptUpdateForm) => void }) {
  const { register, handleSubmit, control, formState: { errors } } = useForm<DeptUpdateForm>({
    resolver: zodResolver(deptUpdateSchema),
    defaultValues: initial,
  });
  return (
    <form id="dept-edit-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label>Mã phòng ban *</Label>
          <Input {...register('code')} />
          {errors.code && <p className="text-xs text-danger">{errors.code.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Trạng thái</Label>
          <Controller
            control={control}
            name="is_active"
            render={({ field }) => (
              <Select value={field.value ? 'true' : 'false'} onValueChange={v => field.onChange(v === 'true')}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">Đang hoạt động</SelectItem>
                  <SelectItem value="false">Ngừng hoạt động</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Tên phòng ban *</Label>
          <Input {...register('name')} />
          {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Chi nhánh</Label>
          <Controller
            control={control}
            name="branch_id"
            render={({ field }) => (
              <Select value={field.value ?? 'none'} onValueChange={v => field.onChange(v === 'none' ? undefined : v)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Không thuộc chi nhánh</SelectItem>
                  {branches.map(b => <SelectItem key={b.id} value={b.id}>{b.name}</SelectItem>)}
                </SelectContent>
              </Select>
            )}
          />
        </div>
      </div>
    </form>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function BranchesPage() {
  const qc = useQueryClient();

  // Branch state
  const [branchPage, setBranchPage] = React.useState(1);
  const [branchQ, setBranchQ] = React.useState('');
  const [branchDialog, setBranchDialog] = React.useState<'create' | 'edit' | null>(null);
  const [selectedBranchId, setSelectedBranchId] = React.useState<string | null>(null);

  // Department state
  const [deptPage, setDeptPage] = React.useState(1);
  const [deptQ, setDeptQ] = React.useState('');
  const [deptBranchFilter, setDeptBranchFilter] = React.useState('');
  const [deptDialog, setDeptDialog] = React.useState<'create' | 'edit' | null>(null);
  const [selectedDeptId, setSelectedDeptId] = React.useState<string | null>(null);

  // Queries
  const branchesQuery = useQuery({
    queryKey: ['branches', branchPage, branchQ],
    queryFn: () => branchService.list({ page: branchPage, size: 20, q: branchQ || undefined }),
  });

  const allBranchesQuery = useQuery({
    queryKey: ['branches-all'],
    queryFn: () => branchService.list({ size: 100 }),
  });

  const deptsQuery = useQuery({
    queryKey: ['departments', deptPage, deptQ, deptBranchFilter],
    queryFn: () => departmentService.list({
      page: deptPage, size: 20,
      q: deptQ || undefined,
      branch_id: deptBranchFilter || undefined,
    }),
  });

  const selectedBranch = branchesQuery.data?.data.find(b => b.id === selectedBranchId);
  const selectedDept = deptsQuery.data?.data.find(d => d.id === selectedDeptId);

  // Branch mutations
  const createBranchMut = useMutation({
    mutationFn: (d: BranchCreateRequest) => branchService.create(d),
    onSuccess: () => {
      toast('Tạo chi nhánh thành công', 'success');
      qc.invalidateQueries({ queryKey: ['branches'] });
      setBranchDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateBranchMut = useMutation({
    mutationFn: (d: BranchUpdateRequest) => branchService.update(selectedBranchId!, d),
    onSuccess: () => {
      toast('Cập nhật chi nhánh thành công', 'success');
      qc.invalidateQueries({ queryKey: ['branches'] });
      setBranchDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  // Department mutations
  const createDeptMut = useMutation({
    mutationFn: (d: DepartmentCreateRequest) => departmentService.create(d),
    onSuccess: () => {
      toast('Tạo phòng ban thành công', 'success');
      qc.invalidateQueries({ queryKey: ['departments'] });
      setDeptDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateDeptMut = useMutation({
    mutationFn: (d: DepartmentUpdateRequest) => departmentService.update(selectedDeptId!, d),
    onSuccess: () => {
      toast('Cập nhật phòng ban thành công', 'success');
      qc.invalidateQueries({ queryKey: ['departments'] });
      setDeptDialog(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const allBranches = allBranchesQuery.data?.data ?? [];

  return (
    <div className="flex flex-col gap-6">
      {/* ── Branches ── */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-4">
            <CardTitle>Chi nhánh</CardTitle>
            <div className="flex items-center gap-2">
              <SearchInput
                placeholder="Tìm theo tên, mã..."
                className="w-60"
                onSearch={v => { setBranchQ(v); setBranchPage(1); }}
              />
              <Button size="sm" onClick={() => setBranchDialog('create')}>
                <Plus className="w-4 h-4" />Thêm chi nhánh
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {branchesQuery.isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Mã</TableHead>
                  <TableHead>Tên chi nhánh</TableHead>
                  <TableHead>Địa chỉ</TableHead>
                  <TableHead>Điện thoại</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="w-16"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {branchesQuery.data?.data.map((branch) => (
                  <TableRow key={branch.id}>
                    <TableCell className="font-mono text-xs">{branch.code}</TableCell>
                    <TableCell className="font-medium">{branch.name}</TableCell>
                    <TableCell className="text-xs text-text-secondary">{branch.address ?? '-'}</TableCell>
                    <TableCell className="text-xs">{branch.phone ?? '-'}</TableCell>
                    <TableCell>
                      <Badge variant={branch.is_active ? 'success' : 'ghost'}>
                        {branch.is_active ? 'Hoạt động' : 'Ngừng'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost" size="icon"
                        onClick={() => { setSelectedBranchId(branch.id); setBranchDialog('edit'); }}
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {branchesQuery.data?.data.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-text-secondary py-8">
                      Chưa có chi nhánh nào
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
        {branchesQuery.data && (
          <div className="px-6 py-3">
            <Pagination page={branchPage} totalPages={branchesQuery.data.total_pages} onPageChange={setBranchPage} />
          </div>
        )}
      </Card>

      {/* ── Departments ── */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-4">
            <CardTitle>Phòng ban</CardTitle>
            <div className="flex items-center gap-2">
              <Select
                value={deptBranchFilter || 'all'}
                onValueChange={v => { setDeptBranchFilter(v === 'all' ? '' : v); setDeptPage(1); }}
              >
                <SelectTrigger className="w-44"><SelectValue placeholder="Tất cả chi nhánh" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Tất cả chi nhánh</SelectItem>
                  {allBranches.map(b => <SelectItem key={b.id} value={b.id}>{b.name}</SelectItem>)}
                </SelectContent>
              </Select>
              <SearchInput
                placeholder="Tìm theo tên, mã..."
                className="w-52"
                onSearch={v => { setDeptQ(v); setDeptPage(1); }}
              />
              <Button size="sm" onClick={() => setDeptDialog('create')}>
                <Plus className="w-4 h-4" />Thêm phòng ban
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {deptsQuery.isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Mã</TableHead>
                  <TableHead>Tên phòng ban</TableHead>
                  <TableHead>Chi nhánh</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="w-16"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {deptsQuery.data?.data.map((dept) => (
                  <TableRow key={dept.id}>
                    <TableCell className="font-mono text-xs">{dept.code}</TableCell>
                    <TableCell className="font-medium">{dept.name}</TableCell>
                    <TableCell className="text-xs text-text-secondary">
                      {allBranches.find(b => b.id === dept.branch_id)?.name ?? '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={dept.is_active ? 'success' : 'ghost'}>
                        {dept.is_active ? 'Hoạt động' : 'Ngừng'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost" size="icon"
                        onClick={() => { setSelectedDeptId(dept.id); setDeptDialog('edit'); }}
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {deptsQuery.data?.data.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-text-secondary py-8">
                      Chưa có phòng ban nào
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
        {deptsQuery.data && (
          <div className="px-6 py-3">
            <Pagination page={deptPage} totalPages={deptsQuery.data.total_pages} onPageChange={setDeptPage} />
          </div>
        )}
      </Card>

      {/* ── Branch Dialogs ── */}
      <Dialog open={branchDialog === 'create'} onOpenChange={o => !o && setBranchDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Thêm chi nhánh mới</DialogTitle></DialogHeader>
          <BranchCreateForm onSubmit={d => createBranchMut.mutate(d)} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setBranchDialog(null)}>Hủy</Button>
            <Button type="submit" form="branch-create-form" loading={createBranchMut.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={branchDialog === 'edit'} onOpenChange={o => !o && setBranchDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Chỉnh sửa chi nhánh</DialogTitle></DialogHeader>
          {selectedBranch && (
            <BranchEditForm
              initial={{
                code: selectedBranch.code,
                name: selectedBranch.name,
                address: selectedBranch.address,
                phone: selectedBranch.phone,
                is_active: selectedBranch.is_active,
              }}
              onSubmit={d => updateBranchMut.mutate(d)}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setBranchDialog(null)}>Hủy</Button>
            <Button type="submit" form="branch-edit-form" loading={updateBranchMut.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Department Dialogs ── */}
      <Dialog open={deptDialog === 'create'} onOpenChange={o => !o && setDeptDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Thêm phòng ban mới</DialogTitle></DialogHeader>
          <DeptCreateForm branches={allBranches} onSubmit={d => createDeptMut.mutate(d)} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeptDialog(null)}>Hủy</Button>
            <Button type="submit" form="dept-create-form" loading={createDeptMut.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deptDialog === 'edit'} onOpenChange={o => !o && setDeptDialog(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>Chỉnh sửa phòng ban</DialogTitle></DialogHeader>
          {selectedDept && (
            <DeptEditForm
              branches={allBranches}
              initial={{
                code: selectedDept.code,
                name: selectedDept.name,
                branch_id: selectedDept.branch_id,
                is_active: selectedDept.is_active,
              }}
              onSubmit={d => updateDeptMut.mutate(d)}
            />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeptDialog(null)}>Hủy</Button>
            <Button type="submit" form="dept-edit-form" loading={updateDeptMut.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
