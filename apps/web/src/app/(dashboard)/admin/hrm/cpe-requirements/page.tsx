'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { MODULE_ROLES, hasAnyRole } from '@/lib/roles';
import {
  cpeRequirementService,
  type CPERequirement,
  type CreateCPERequirementRequest,
  type UpdateCPERequirementRequest,
} from '@/services/hrm/professional';
import { Plus, Pencil, Award } from 'lucide-react';

const INTERNAL_ROLES = [
  'SUPER_ADMIN', 'CHAIRMAN', 'CEO', 'FIRM_PARTNER',
  'AUDIT_MANAGER', 'AUDIT_STAFF', 'HR_MANAGER', 'HR_STAFF',
  'HEAD_OF_BRANCH', 'ACCOUNTANT',
];

const ROLE_LABELS: Record<string, string> = {
  SUPER_ADMIN:    'Quản trị viên',
  CHAIRMAN:       'Chủ tịch HĐTV',
  CEO:            'Giám đốc',
  FIRM_PARTNER:   'Partner',
  AUDIT_MANAGER:  'Trưởng nhóm KT',
  AUDIT_STAFF:    'Kiểm toán viên',
  HR_MANAGER:     'Trưởng nhân sự',
  HR_STAFF:       'Nhân viên NS',
  HEAD_OF_BRANCH: 'Trưởng chi nhánh',
  ACCOUNTANT:     'Kế toán',
};

const CURRENT_YEAR = new Date().getFullYear();
const PAGE_SIZE = 50;

export default function CPERequirementsPage() {
  const { user } = useAuthStore();
  const qc = useQueryClient();

  const canWrite = hasAnyRole(user?.roles ?? [], MODULE_ROLES.hrmCPEWrite);

  const [filterYear, setFilterYear] = React.useState<number>(CURRENT_YEAR);
  const [filterRole, setFilterRole] = React.useState('');

  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<CPERequirement | null>(null);

  const [createForm, setCreateForm] = React.useState({
    role_code: INTERNAL_ROLES[0],
    year: String(CURRENT_YEAR),
    required_hours: '',
    notes: '',
  });
  const [editForm, setEditForm] = React.useState({ required_hours: '', notes: '' });

  const queryKey = ['hrm', 'cpe-requirements', filterYear, filterRole];

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => cpeRequirementService.list({
      page: 1,
      size: PAGE_SIZE,
      year: filterYear,
      role_code: filterRole || undefined,
    }),
  });

  const createMutation = useMutation({
    mutationFn: (req: CreateCPERequirementRequest) => cpeRequirementService.create(req),
    onSuccess: () => {
      toast('Tạo yêu cầu CPE thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'cpe-requirements'] });
      setCreateOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateCPERequirementRequest }) =>
      cpeRequirementService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật yêu cầu CPE thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'cpe-requirements'] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  function openEdit(req: CPERequirement) {
    setEditForm({
      required_hours: String(req.required_hours),
      notes: req.notes ?? '',
    });
    setEditTarget(req);
  }

  function submitCreate(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate({
      role_code: createForm.role_code,
      year: parseInt(createForm.year, 10),
      required_hours: parseFloat(createForm.required_hours) || 0,
      notes: createForm.notes || undefined,
    });
  }

  function submitEdit(e: React.FormEvent) {
    e.preventDefault();
    if (!editTarget) return;
    updateMutation.mutate({
      id: editTarget.id,
      data: {
        required_hours: parseFloat(editForm.required_hours) || 0,
        notes: editForm.notes || undefined,
      },
    });
  }

  const items = data?.data ?? [];

  const yearOptions = Array.from({ length: 5 }, (_, i) => CURRENT_YEAR - 2 + i);

  return (
    <div className="flex flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Yêu cầu CPE theo vai trò</h1>
          <p className="text-sm text-text-secondary">Cấu hình số giờ CPE bắt buộc mỗi năm</p>
        </div>
        {canWrite && (
          <Button onClick={() => { setCreateForm({ role_code: INTERNAL_ROLES[0], year: String(CURRENT_YEAR), required_hours: '', notes: '' }); setCreateOpen(true); }}>
            <Plus className="w-4 h-4" />Thêm yêu cầu
          </Button>
        )}
      </div>

      {/* Filters */}
      <div className="flex gap-2 flex-wrap">
        <select
          className="text-sm border rounded px-3 py-1.5 bg-bg-primary text-text-primary"
          value={filterYear}
          onChange={e => setFilterYear(Number(e.target.value))}
        >
          {yearOptions.map(y => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
        <select
          className="text-sm border rounded px-3 py-1.5 bg-bg-primary text-text-primary"
          value={filterRole}
          onChange={e => setFilterRole(e.target.value)}
        >
          <option value="">Tất cả vai trò</option>
          {INTERNAL_ROLES.map(r => (
            <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <PageSpinner />
          ) : items.length === 0 ? (
            <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
              <Award className="w-10 h-10 opacity-30" />
              <p>Chưa có yêu cầu CPE nào cho bộ lọc này.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-bg-secondary">
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Vai trò</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Năm</th>
                    <th className="text-right px-4 py-3 text-text-secondary font-medium">Số giờ yêu cầu</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Ghi chú</th>
                    {canWrite && <th className="px-4 py-3" />}
                  </tr>
                </thead>
                <tbody>
                  {items.map(req => (
                    <tr key={req.id} className="border-b hover:bg-bg-secondary/50 transition-colors">
                      <td className="px-4 py-3 font-medium text-text-primary">
                        {ROLE_LABELS[req.role_code] ?? req.role_code}
                      </td>
                      <td className="px-4 py-3 tabular-nums">{req.year}</td>
                      <td className="px-4 py-3 text-right tabular-nums font-semibold">
                        {req.required_hours} giờ
                      </td>
                      <td className="px-4 py-3 text-text-secondary">{req.notes ?? '–'}</td>
                      {canWrite && (
                        <td className="px-4 py-3">
                          <Button variant="ghost" size="icon" onClick={() => openEdit(req)}>
                            <Pencil className="w-4 h-4" />
                          </Button>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Thêm yêu cầu CPE</DialogTitle></DialogHeader>
          <form id="cpe-create-form" onSubmit={submitCreate} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Vai trò <span className="text-danger">*</span></Label>
              <select
                className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                value={createForm.role_code}
                onChange={e => setCreateForm(f => ({ ...f, role_code: e.target.value }))}
              >
                {INTERNAL_ROLES.map(r => (
                  <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Năm <span className="text-danger">*</span></Label>
              <Input
                type="number"
                min={2000}
                max={2100}
                value={createForm.year}
                onChange={e => setCreateForm(f => ({ ...f, year: e.target.value }))}
                required
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số giờ CPE yêu cầu <span className="text-danger">*</span></Label>
              <Input
                type="number"
                min={0}
                step={0.5}
                value={createForm.required_hours}
                onChange={e => setCreateForm(f => ({ ...f, required_hours: e.target.value }))}
                required
                placeholder="VD: 40"
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ghi chú</Label>
              <Input
                value={createForm.notes}
                onChange={e => setCreateForm(f => ({ ...f, notes: e.target.value }))}
              />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="cpe-create-form" loading={createMutation.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      <Dialog open={!!editTarget} onOpenChange={v => !v && setEditTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>
              Cập nhật CPE — {editTarget ? ROLE_LABELS[editTarget.role_code] ?? editTarget.role_code : ''} {editTarget?.year}
            </DialogTitle>
          </DialogHeader>
          <form id="cpe-edit-form" onSubmit={submitEdit} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Số giờ CPE yêu cầu <span className="text-danger">*</span></Label>
              <Input
                type="number"
                min={0}
                step={0.5}
                value={editForm.required_hours}
                onChange={e => setEditForm(f => ({ ...f, required_hours: e.target.value }))}
                required
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ghi chú</Label>
              <Input
                value={editForm.notes}
                onChange={e => setEditForm(f => ({ ...f, notes: e.target.value }))}
              />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTarget(null)}>Hủy</Button>
            <Button type="submit" form="cpe-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
