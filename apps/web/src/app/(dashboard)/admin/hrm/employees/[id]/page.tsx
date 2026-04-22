'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { UpdateEmployeeForm } from '@/components/hrm/employee-form';
import { DependentSection } from '@/components/hrm/dependent-section';
import { ContractSection } from '@/components/hrm/contract-section';
import { SensitiveModal } from '@/components/hrm/sensitive-modal';
import { SalaryHistorySection } from '@/components/hrm/salary-history-section';
import { employeeService, type UpdateEmployeeRequest } from '@/services/hrm/employee';
import {
  certificationService,
  trainingRecordService,
  trainingCourseService,
  type Certification,
  type TrainingRecord,
  type TrainingCourse,
  type CertType,
  type CertStatus,
  type CourseType,
  type TrainingStatus,
} from '@/services/hrm/professional';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { ArrowLeft, Pencil, ShieldAlert, Plus, Trash2, Award, BookOpen } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'HR_MANAGER'];
const SENSITIVE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO', 'HR_MANAGER'];
const SALARY_READ_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO', 'HR_MANAGER'];
const SALARY_WRITE_ROLES = ['SUPER_ADMIN', 'CEO', 'HR_MANAGER'];

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  ACTIVE: 'default', INACTIVE: 'secondary', ON_LEAVE: 'outline', TERMINATED: 'secondary',
};
const STATUS_LABELS: Record<string, string> = {
  ACTIVE: 'Đang làm', INACTIVE: 'Không HĐ', ON_LEAVE: 'Nghỉ phép', TERMINATED: 'Đã nghỉ',
};

const CERT_TYPE_LABELS: Record<CertType, string> = {
  VN_CPA: 'CPA VN', ACCA: 'ACCA', CPA_AUSTRALIA: 'CPA Úc', CFA: 'CFA',
  CIA: 'CIA', CISA: 'CISA', IFRS: 'IFRS', ICAEW: 'ICAEW', CMA: 'CMA', OTHER: 'Khác',
};
const CERT_STATUS_LABELS: Record<CertStatus, string> = {
  ACTIVE: 'Còn hạn', EXPIRED: 'Hết hạn', REVOKED: 'Đã thu hồi', SUSPENDED: 'Tạm đình chỉ',
};
const CERT_STATUS_VARIANTS: Record<CertStatus, 'default' | 'secondary' | 'outline'> = {
  ACTIVE: 'default', EXPIRED: 'secondary', REVOKED: 'secondary', SUSPENDED: 'outline',
};
const COURSE_TYPE_LABELS: Record<CourseType, string> = {
  TECHNICAL: 'Chuyên môn', ETHICS: 'Đạo đức', MANAGEMENT: 'Quản lý',
  SOFT_SKILLS: 'Kỹ năng mềm', COMPLIANCE: 'Tuân thủ', OTHER: 'Khác',
};
const TRAINING_STATUS_LABELS: Record<TrainingStatus, string> = {
  ENROLLED: 'Đã đăng ký', IN_PROGRESS: 'Đang học',
  COMPLETED: 'Hoàn thành', FAILED: 'Không đạt', CANCELLED: 'Đã hủy',
};
const TRAINING_STATUS_VARIANTS: Record<TrainingStatus, 'default' | 'secondary' | 'outline'> = {
  ENROLLED: 'outline', IN_PROGRESS: 'outline', COMPLETED: 'default',
  FAILED: 'secondary', CANCELLED: 'secondary',
};

const CERT_TYPES: CertType[] = ['VN_CPA','ACCA','CPA_AUSTRALIA','CFA','CIA','CISA','IFRS','ICAEW','CMA','OTHER'];
const CERT_STATUSES: CertStatus[] = ['ACTIVE','EXPIRED','REVOKED','SUSPENDED'];
const TRAINING_STATUSES: TrainingStatus[] = ['ENROLLED','IN_PROGRESS','COMPLETED','FAILED','CANCELLED'];

type TabKey = 'info' | 'dependents' | 'contracts' | 'salary' | 'certifications' | 'training';

function InfoRow({ label, value }: { label: string; value?: string | number | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value ?? '–'}</span>
    </div>
  );
}

function TabButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? 'border-action text-action'
          : 'border-transparent text-text-secondary hover:text-text-primary'
      }`}
    >
      {children}
    </button>
  );
}

// ─── Certification Section ─────────────────────────────────────────────────────

interface CertFormState {
  cert_type: CertType;
  cert_number: string;
  issued_date: string;
  expiry_date: string;
  issuing_authority: string;
  status: CertStatus;
  notes: string;
}
const DEFAULT_CERT: CertFormState = {
  cert_type: 'VN_CPA', cert_number: '', issued_date: '',
  expiry_date: '', issuing_authority: '', status: 'ACTIVE', notes: '',
};

function CertificationSection({ employeeId, canWrite }: { employeeId: string; canWrite: boolean }) {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<Certification | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Certification | null>(null);
  const [form, setForm] = React.useState<CertFormState>(DEFAULT_CERT);

  const { data, isLoading } = useQuery({
    queryKey: ['hrm', 'employees', employeeId, 'certifications'],
    queryFn: () => certificationService.listByEmployee(employeeId, { size: 50 }),
  });

  const createMutation = useMutation({
    mutationFn: () => certificationService.create(employeeId, {
      cert_type: form.cert_type,
      cert_name: CERT_TYPE_LABELS[form.cert_type],
      cert_number: form.cert_number,
      issued_date: form.issued_date,
      expiry_date: form.expiry_date || undefined,
      issuing_authority: form.issuing_authority,
      status: form.status,
      notes: form.notes || undefined,
    }),
    onSuccess: () => {
      toast('Thêm chứng chỉ thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'certifications'] });
      setCreateOpen(false);
      setForm(DEFAULT_CERT);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: () => certificationService.update(editTarget!.id, {
      cert_type: form.cert_type,
      cert_name: CERT_TYPE_LABELS[form.cert_type],
      cert_number: form.cert_number,
      issued_date: form.issued_date,
      expiry_date: form.expiry_date || undefined,
      issuing_authority: form.issuing_authority,
      status: form.status,
      notes: form.notes || undefined,
    }),
    onSuccess: () => {
      toast('Cập nhật chứng chỉ thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'certifications'] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => certificationService.delete(deleteTarget!.id),
    onSuccess: () => {
      toast('Xóa chứng chỉ thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'certifications'] });
      setDeleteTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  function openEdit(cert: Certification) {
    setForm({
      cert_type: cert.cert_type,
      cert_number: cert.cert_number,
      issued_date: cert.issued_date,
      expiry_date: cert.expiry_date ?? '',
      issuing_authority: cert.issuing_authority,
      status: cert.status,
      notes: cert.notes ?? '',
    });
    setEditTarget(cert);
  }

  const certs = data?.data ?? [];

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <p className="font-medium text-text-primary">Chứng chỉ chuyên môn</p>
        {canWrite && (
          <Button size="sm" onClick={() => { setForm(DEFAULT_CERT); setCreateOpen(true); }}>
            <Plus className="w-3 h-3" />Thêm chứng chỉ
          </Button>
        )}
      </div>

      {isLoading ? (
        <PageSpinner />
      ) : certs.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-8 text-text-secondary">
          <Award className="w-8 h-8 opacity-30" />
          <p className="text-sm">Chưa có chứng chỉ nào.</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-bg-secondary">
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Loại</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Số hiệu</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Cơ quan cấp</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Ngày cấp</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Hết hạn</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Trạng thái</th>
                {canWrite && <th className="px-3 py-2" />}
              </tr>
            </thead>
            <tbody>
              {certs.map(cert => (
                <tr key={cert.id} className="border-b hover:bg-bg-secondary/50">
                  <td className="px-3 py-2">
                    <Badge variant="outline">{CERT_TYPE_LABELS[cert.cert_type]}</Badge>
                  </td>
                  <td className="px-3 py-2 font-mono text-xs">{cert.cert_number}</td>
                  <td className="px-3 py-2 text-text-secondary">{cert.issuing_authority}</td>
                  <td className="px-3 py-2">{formatDate(cert.issued_date)}</td>
                  <td className="px-3 py-2">{cert.expiry_date ? formatDate(cert.expiry_date) : '–'}</td>
                  <td className="px-3 py-2">
                    <Badge variant={CERT_STATUS_VARIANTS[cert.status]}>
                      {CERT_STATUS_LABELS[cert.status]}
                    </Badge>
                  </td>
                  {canWrite && (
                    <td className="px-3 py-2">
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" onClick={() => openEdit(cert)}>
                          <Pencil className="w-3 h-3" />
                        </Button>
                        <Button
                          variant="ghost" size="icon"
                          className="text-danger hover:text-danger"
                          onClick={() => setDeleteTarget(cert)}
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create / Edit dialog */}
      {[
        { open: createOpen, onClose: () => setCreateOpen(false), title: 'Thêm chứng chỉ', formId: 'cert-create', onSubmit: (e: React.FormEvent) => { e.preventDefault(); createMutation.mutate(); }, isPending: createMutation.isPending },
        { open: !!editTarget, onClose: () => setEditTarget(null), title: 'Chỉnh sửa chứng chỉ', formId: 'cert-edit', onSubmit: (e: React.FormEvent) => { e.preventDefault(); updateMutation.mutate(); }, isPending: updateMutation.isPending },
      ].map(({ open, onClose, title, formId, onSubmit, isPending }) => (
        <Dialog key={formId} open={open} onOpenChange={v => !v && onClose()}>
          <DialogContent className="max-w-md">
            <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
            <form id={formId} onSubmit={onSubmit} className="flex flex-col gap-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <Label>Loại chứng chỉ</Label>
                  <select
                    className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                    value={form.cert_type}
                    onChange={e => setForm(f => ({ ...f, cert_type: e.target.value as CertType }))}
                  >
                    {CERT_TYPES.map(t => <option key={t} value={t}>{CERT_TYPE_LABELS[t]}</option>)}
                  </select>
                </div>
                <div className="flex flex-col gap-1">
                  <Label>Trạng thái</Label>
                  <select
                    className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                    value={form.status}
                    onChange={e => setForm(f => ({ ...f, status: e.target.value as CertStatus }))}
                  >
                    {CERT_STATUSES.map(s => <option key={s} value={s}>{CERT_STATUS_LABELS[s]}</option>)}
                  </select>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label>Số hiệu chứng chỉ <span className="text-danger">*</span></Label>
                <Input value={form.cert_number} onChange={e => setForm(f => ({ ...f, cert_number: e.target.value }))} required />
              </div>
              <div className="flex flex-col gap-1">
                <Label>Cơ quan cấp <span className="text-danger">*</span></Label>
                <Input value={form.issuing_authority} onChange={e => setForm(f => ({ ...f, issuing_authority: e.target.value }))} required />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <Label>Ngày cấp <span className="text-danger">*</span></Label>
                  <Input type="date" value={form.issued_date} onChange={e => setForm(f => ({ ...f, issued_date: e.target.value }))} required />
                </div>
                <div className="flex flex-col gap-1">
                  <Label>Ngày hết hạn</Label>
                  <Input type="date" value={form.expiry_date} onChange={e => setForm(f => ({ ...f, expiry_date: e.target.value }))} />
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <Label>Ghi chú</Label>
                <Input value={form.notes} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} />
              </div>
            </form>
            <DialogFooter>
              <Button variant="outline" onClick={onClose}>Hủy</Button>
              <Button type="submit" form={formId} loading={isPending}>Lưu</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ))}

      {/* Delete confirm */}
      <Dialog open={!!deleteTarget} onOpenChange={v => !v && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Xóa chứng chỉ?</DialogTitle></DialogHeader>
          <p className="text-sm text-text-secondary">
            Xóa chứng chỉ <strong>{deleteTarget ? CERT_TYPE_LABELS[deleteTarget.cert_type] : ''}</strong> — {deleteTarget?.cert_number}? Hành động này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>Hủy</Button>
            <Button variant="danger" loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate()}>Xóa</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ─── Training Records Section ──────────────────────────────────────────────────

interface TrainingFormState {
  course_id: string;
  completion_date: string;
  status: TrainingStatus;
  cpe_hours_earned: string;
  notes: string;
}
const DEFAULT_TRAINING: TrainingFormState = {
  course_id: '', completion_date: '',
  status: 'ENROLLED', cpe_hours_earned: '0', notes: '',
};

function TrainingSection({ employeeId, canWrite }: { employeeId: string; canWrite: boolean }) {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<TrainingRecord | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<TrainingRecord | null>(null);
  const [form, setForm] = React.useState<TrainingFormState>(DEFAULT_TRAINING);

  const { data, isLoading } = useQuery({
    queryKey: ['hrm', 'employees', employeeId, 'training-records'],
    queryFn: () => trainingRecordService.listByEmployee(employeeId, { size: 50 }),
  });

  const { data: coursesData } = useQuery({
    queryKey: ['hrm', 'training-courses', 'all'],
    queryFn: () => trainingCourseService.list({ size: 200, is_active: true }),
    enabled: createOpen,
  });
  const courses: TrainingCourse[] = coursesData?.data ?? [];

  const createMutation = useMutation({
    mutationFn: () => trainingRecordService.create(employeeId, {
      course_id: form.course_id,
      status: form.status,
      cpe_hours_earned: parseFloat(form.cpe_hours_earned) || 0,
      notes: form.notes || undefined,
    }),
    onSuccess: () => {
      toast('Thêm hồ sơ đào tạo thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'training-records'] });
      setCreateOpen(false);
      setForm(DEFAULT_TRAINING);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: () => trainingRecordService.update(editTarget!.id, {
      completion_date: form.completion_date || undefined,
      status: form.status,
      cpe_hours_earned: parseFloat(form.cpe_hours_earned) || 0,
      notes: form.notes || undefined,
    }),
    onSuccess: () => {
      toast('Cập nhật hồ sơ đào tạo thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'training-records'] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => trainingRecordService.delete(deleteTarget!.id),
    onSuccess: () => {
      toast('Xóa hồ sơ đào tạo thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', employeeId, 'training-records'] });
      setDeleteTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  function openEdit(rec: TrainingRecord) {
    setForm({
      course_id: rec.course_id,
      completion_date: rec.completion_date ?? '',
      status: rec.status,
      cpe_hours_earned: String(rec.cpe_hours_earned),
      notes: rec.notes ?? '',
    });
    setEditTarget(rec);
  }

  const records = data?.data ?? [];

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <p className="font-medium text-text-primary">Hồ sơ đào tạo</p>
        {canWrite && (
          <Button size="sm" onClick={() => { setForm(DEFAULT_TRAINING); setCreateOpen(true); }}>
            <Plus className="w-3 h-3" />Thêm hồ sơ
          </Button>
        )}
      </div>

      {isLoading ? (
        <PageSpinner />
      ) : records.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-8 text-text-secondary">
          <BookOpen className="w-8 h-8 opacity-30" />
          <p className="text-sm">Chưa có hồ sơ đào tạo nào.</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-bg-secondary">
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Khóa học</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Loại</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Trạng thái</th>
                <th className="text-right px-3 py-2 text-text-secondary font-medium">Giờ CPE</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Ngày đăng ký</th>
                <th className="text-left px-3 py-2 text-text-secondary font-medium">Hoàn thành</th>
                {canWrite && <th className="px-3 py-2" />}
              </tr>
            </thead>
            <tbody>
              {records.map(rec => (
                <tr key={rec.id} className="border-b hover:bg-bg-secondary/50">
                  <td className="px-3 py-2 font-medium text-text-primary">
                    {rec.course_name ?? rec.course_id}
                  </td>
                  <td className="px-3 py-2">
                    {rec.course_type ? (
                      <Badge variant="outline">{COURSE_TYPE_LABELS[rec.course_type]}</Badge>
                    ) : '–'}
                  </td>
                  <td className="px-3 py-2">
                    <Badge variant={TRAINING_STATUS_VARIANTS[rec.status]}>
                      {TRAINING_STATUS_LABELS[rec.status]}
                    </Badge>
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums">{rec.cpe_hours_earned}</td>
                  <td className="px-3 py-2">{formatDate(rec.created_at)}</td>
                  <td className="px-3 py-2">{rec.completion_date ? formatDate(rec.completion_date) : '–'}</td>
                  {canWrite && (
                    <td className="px-3 py-2">
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" onClick={() => openEdit(rec)}>
                          <Pencil className="w-3 h-3" />
                        </Button>
                        <Button
                          variant="ghost" size="icon"
                          className="text-danger hover:text-danger"
                          onClick={() => setDeleteTarget(rec)}
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={v => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Thêm hồ sơ đào tạo</DialogTitle></DialogHeader>
          <form id="tr-create" onSubmit={e => { e.preventDefault(); createMutation.mutate(); }} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Khóa học <span className="text-danger">*</span></Label>
              <select
                className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                value={form.course_id}
                onChange={e => setForm(f => ({ ...f, course_id: e.target.value }))}
                required
              >
                <option value="">-- Chọn khóa học --</option>
                {courses.map(c => (
                  <option key={c.id} value={c.id}>{c.name} ({c.code})</option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Trạng thái</Label>
              <select
                className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value as TrainingStatus }))}
              >
                {TRAINING_STATUSES.map(s => <option key={s} value={s}>{TRAINING_STATUS_LABELS[s]}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số giờ CPE đạt được</Label>
              <Input type="number" min={0} step={0.5} value={form.cpe_hours_earned} onChange={e => setForm(f => ({ ...f, cpe_hours_earned: e.target.value }))} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ghi chú</Label>
              <Input value={form.notes} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="tr-create" loading={createMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      <Dialog open={!!editTarget} onOpenChange={v => !v && setEditTarget(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Cập nhật hồ sơ đào tạo</DialogTitle></DialogHeader>
          <form id="tr-edit" onSubmit={e => { e.preventDefault(); updateMutation.mutate(); }} className="flex flex-col gap-3">
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1">
                <Label>Trạng thái</Label>
                <select
                  className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
                  value={form.status}
                  onChange={e => setForm(f => ({ ...f, status: e.target.value as TrainingStatus }))}
                >
                  {TRAINING_STATUSES.map(s => <option key={s} value={s}>{TRAINING_STATUS_LABELS[s]}</option>)}
                </select>
              </div>
              <div className="flex flex-col gap-1">
                <Label>Ngày hoàn thành</Label>
                <Input type="date" value={form.completion_date} onChange={e => setForm(f => ({ ...f, completion_date: e.target.value }))} />
              </div>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số giờ CPE đạt được</Label>
              <Input type="number" min={0} step={0.5} value={form.cpe_hours_earned} onChange={e => setForm(f => ({ ...f, cpe_hours_earned: e.target.value }))} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ghi chú</Label>
              <Input value={form.notes} onChange={e => setForm(f => ({ ...f, notes: e.target.value }))} />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTarget(null)}>Hủy</Button>
            <Button type="submit" form="tr-edit" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm */}
      <Dialog open={!!deleteTarget} onOpenChange={v => !v && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Xóa hồ sơ đào tạo?</DialogTitle></DialogHeader>
          <p className="text-sm text-text-secondary">
            Xóa hồ sơ khóa học <strong>{deleteTarget?.course_name ?? deleteTarget?.course_id}</strong>? Hành động này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>Hủy</Button>
            <Button variant="danger" loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate()}>Xóa</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────────

export default function EmployeeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [tab, setTab] = React.useState<TabKey>('info');
  const [editOpen, setEditOpen] = React.useState(false);
  const [sensitiveOpen, setSensitiveOpen] = React.useState(false);

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));
  const canSeeSensitive = SENSITIVE_ROLES.some(r => user?.roles?.includes(r));
  const canReadSalary = SALARY_READ_ROLES.some(r => user?.roles?.includes(r));
  const canWriteSalary = SALARY_WRITE_ROLES.some(r => user?.roles?.includes(r));
  const canWriteDependent = [...WRITE_ROLES, 'HR_STAFF'].some(r => user?.roles?.includes(r));

  const { data: emp, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'employees', id],
    queryFn: () => employeeService.getById(id),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateEmployeeRequest) => employeeService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật nhân viên thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', id] });
      setEditOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError || !emp) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không tìm thấy nhân viên.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 max-w-4xl">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/employees')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h1 className="text-xl font-semibold text-text-primary">{emp.full_name}</h1>
            <Badge variant={STATUS_VARIANTS[emp.status] ?? 'outline'}>
              {STATUS_LABELS[emp.status] ?? emp.status}
            </Badge>
            <Badge variant="outline">{emp.grade}</Badge>
          </div>
          <p className="text-xs text-text-secondary font-mono mt-0.5">{emp.email}</p>
        </div>
        <div className="flex gap-2">
          {canSeeSensitive && (
            <Button variant="outline" className="text-danger border-danger hover:bg-danger/5" onClick={() => setSensitiveOpen(true)}>
              <ShieldAlert className="w-4 h-4" />PII
            </Button>
          )}
          {canWrite && (
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="w-4 h-4" />Chỉnh sửa
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b flex gap-0 flex-wrap">
        <TabButton active={tab === 'info'} onClick={() => setTab('info')}>Thông tin</TabButton>
        <TabButton active={tab === 'dependents'} onClick={() => setTab('dependents')}>Người phụ thuộc</TabButton>
        <TabButton active={tab === 'contracts'} onClick={() => setTab('contracts')}>Hợp đồng</TabButton>
        {canReadSalary && (
          <TabButton active={tab === 'salary'} onClick={() => setTab('salary')}>Lịch sử lương</TabButton>
        )}
        <TabButton active={tab === 'certifications'} onClick={() => setTab('certifications')}>Chứng chỉ</TabButton>
        <TabButton active={tab === 'training'} onClick={() => setTab('training')}>Đào tạo</TabButton>
      </div>

      {/* Tab: Thông tin */}
      {tab === 'info' && (
        <Card>
          <CardContent className="p-5">
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Mã nhân viên" value={emp.employee_code} />
              <InfoRow label="Tên hiển thị" value={emp.display_name} />
              <InfoRow label="Điện thoại" value={emp.phone} />
              <InfoRow label="Ngày sinh" value={emp.date_of_birth} />
              <InfoRow label="Giới tính" value={emp.gender} />
              <InfoRow label="Quốc tịch" value={emp.nationality} />
              <InfoRow label="Chức vụ" value={emp.position_title} />
              <InfoRow label="Địa điểm làm việc" value={emp.work_location} />
              <InfoRow label="Ngày bắt đầu" value={emp.hired_date} />
              <InfoRow label="Ngày kết thúc thử việc" value={emp.probation_end_date} />
              <InfoRow label="Email cá nhân" value={emp.personal_email} />
              <InfoRow label="ĐT cá nhân" value={emp.personal_phone} />
              <InfoRow label="ĐT công ty" value={emp.work_phone} />
              <InfoRow label="Trình độ học vấn" value={emp.education_level} />
              <InfoRow label="Chuyên ngành" value={emp.education_major} />
              <InfoRow label="Trường học" value={emp.education_school} />
              <InfoRow label="Loại hoa hồng" value={emp.commission_type} />
              <InfoRow label="Số CPA" value={emp.vn_cpa_number} />
              <InfoRow label="Ngày cập nhật" value={formatDate(emp.updated_at)} />
              <InfoRow label="Ngày tạo" value={formatDate(emp.created_at)} />
              <div className="col-span-2">
                <InfoRow label="Địa chỉ hiện tại" value={emp.current_address} />
              </div>
              <div className="col-span-2">
                <InfoRow label="Địa chỉ thường trú" value={emp.permanent_address} />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tab: Người phụ thuộc */}
      {tab === 'dependents' && (
        <Card>
          <CardContent className="p-5">
            <DependentSection employeeId={id} canWrite={canWriteDependent} />
          </CardContent>
        </Card>
      )}

      {/* Tab: Hợp đồng */}
      {tab === 'contracts' && (
        <Card>
          <CardContent className="p-5">
            <ContractSection employeeId={id} canWrite={canWrite} />
          </CardContent>
        </Card>
      )}

      {/* Tab: Lịch sử lương */}
      {tab === 'salary' && canReadSalary && (
        <Card>
          <CardContent className="p-5">
            <SalaryHistorySection
              employeeId={id}
              canRead={canReadSalary}
              canWrite={canWriteSalary}
            />
          </CardContent>
        </Card>
      )}

      {/* Tab: Chứng chỉ */}
      {tab === 'certifications' && (
        <Card>
          <CardContent className="p-5">
            <CertificationSection employeeId={id} canWrite={canWrite} />
          </CardContent>
        </Card>
      )}

      {/* Tab: Đào tạo */}
      {tab === 'training' && (
        <Card>
          <CardContent className="p-5">
            <TrainingSection employeeId={id} canWrite={canWrite} />
          </CardContent>
        </Card>
      )}

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Chỉnh sửa nhân viên</DialogTitle></DialogHeader>
          <UpdateEmployeeForm
            formId="emp-edit-form"
            initial={emp}
            onSubmit={(data) => updateMutation.mutate(data)}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Hủy</Button>
            <Button type="submit" form="emp-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Sensitive PII modal */}
      {canSeeSensitive && (
        <SensitiveModal
          open={sensitiveOpen}
          onOpenChange={setSensitiveOpen}
          employeeId={id}
          employeeName={emp.full_name}
        />
      )}
    </div>
  );
}
