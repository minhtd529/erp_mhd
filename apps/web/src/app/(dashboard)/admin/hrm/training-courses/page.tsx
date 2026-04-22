'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { MODULE_ROLES, hasAnyRole } from '@/lib/roles';
import {
  trainingCourseService,
  type TrainingCourse,
  type CreateTrainingCourseRequest,
  type UpdateTrainingCourseRequest,
  type CourseType,
} from '@/services/hrm/professional';
import { Plus, Pencil, Trash2, BookOpen } from 'lucide-react';

const COURSE_TYPE_LABELS: Record<CourseType, string> = {
  TECHNICAL:   'Chuyên môn',
  ETHICS:      'Đạo đức nghề nghiệp',
  MANAGEMENT:  'Quản lý',
  SOFT_SKILLS: 'Kỹ năng mềm',
  COMPLIANCE:  'Tuân thủ',
  OTHER:       'Khác',
};

const COURSE_TYPES: CourseType[] = ['TECHNICAL', 'ETHICS', 'MANAGEMENT', 'SOFT_SKILLS', 'COMPLIANCE', 'OTHER'];

const PAGE_SIZE = 20;

interface CourseFormState {
  code: string;
  name: string;
  description: string;
  course_type: CourseType;
  cpe_hours: string;
  provider: string;
  is_active: boolean;
}

const DEFAULT_FORM: CourseFormState = {
  code: '', name: '', description: '', course_type: 'TECHNICAL',
  cpe_hours: '0', provider: '', is_active: true,
};

export default function TrainingCoursesPage() {
  const { user } = useAuthStore();
  const qc = useQueryClient();

  const canWrite  = hasAnyRole(user?.roles ?? [], MODULE_ROLES.hrmTrainingCourseWrite);
  const canDelete = hasAnyRole(user?.roles ?? [], MODULE_ROLES.hrmTrainingCourseDelete);

  const [page, setPage] = React.useState(1);
  const [filterType, setFilterType] = React.useState<CourseType | ''>('');
  const [filterActive, setFilterActive] = React.useState<'' | 'true' | 'false'>('');

  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<TrainingCourse | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<TrainingCourse | null>(null);
  const [form, setForm] = React.useState<CourseFormState>(DEFAULT_FORM);

  const queryKey = ['hrm', 'training-courses', page, filterType, filterActive];

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => trainingCourseService.list({
      page,
      size: PAGE_SIZE,
      course_type: filterType || undefined,
      is_active: filterActive === '' ? undefined : filterActive === 'true',
    }),
  });

  const createMutation = useMutation({
    mutationFn: (req: CreateTrainingCourseRequest) => trainingCourseService.create(req),
    onSuccess: () => {
      toast('Tạo khóa học thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'training-courses'] });
      setCreateOpen(false);
      setForm(DEFAULT_FORM);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTrainingCourseRequest }) =>
      trainingCourseService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật khóa học thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'training-courses'] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => trainingCourseService.delete(id),
    onSuccess: () => {
      toast('Xóa khóa học thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'training-courses'] });
      setDeleteTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  function openEdit(course: TrainingCourse) {
    setForm({
      code: course.code,
      name: course.name,
      description: course.description ?? '',
      course_type: course.course_type,
      cpe_hours: String(course.cpe_hours),
      provider: course.provider ?? '',
      is_active: course.is_active,
    });
    setEditTarget(course);
  }

  function submitCreate(e: React.FormEvent) {
    e.preventDefault();
    createMutation.mutate({
      code: form.code,
      name: form.name,
      description: form.description || undefined,
      course_type: form.course_type,
      cpe_hours: parseFloat(form.cpe_hours) || 0,
      provider: form.provider || undefined,
      is_active: form.is_active,
    });
  }

  function submitEdit(e: React.FormEvent) {
    e.preventDefault();
    if (!editTarget) return;
    updateMutation.mutate({
      id: editTarget.id,
      data: {
        name: form.name || undefined,
        description: form.description || undefined,
        course_type: form.course_type,
        cpe_hours: parseFloat(form.cpe_hours) || 0,
        provider: form.provider || undefined,
        is_active: form.is_active,
      },
    });
  }

  const courses = data?.data ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <div className="flex flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Danh mục khóa học</h1>
          <p className="text-sm text-text-secondary">Quản lý khóa học đào tạo và số giờ CPE</p>
        </div>
        {canWrite && (
          <Button onClick={() => { setForm(DEFAULT_FORM); setCreateOpen(true); }}>
            <Plus className="w-4 h-4" />Thêm khóa học
          </Button>
        )}
      </div>

      {/* Filters */}
      <div className="flex gap-2 flex-wrap">
        <select
          className="text-sm border rounded px-3 py-1.5 bg-bg-primary text-text-primary"
          value={filterType}
          onChange={e => { setFilterType(e.target.value as CourseType | ''); setPage(1); }}
        >
          <option value="">Tất cả loại</option>
          {COURSE_TYPES.map(t => (
            <option key={t} value={t}>{COURSE_TYPE_LABELS[t]}</option>
          ))}
        </select>
        <select
          className="text-sm border rounded px-3 py-1.5 bg-bg-primary text-text-primary"
          value={filterActive}
          onChange={e => { setFilterActive(e.target.value as '' | 'true' | 'false'); setPage(1); }}
        >
          <option value="">Tất cả trạng thái</option>
          <option value="true">Đang hoạt động</option>
          <option value="false">Ngừng hoạt động</option>
        </select>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <PageSpinner />
          ) : courses.length === 0 ? (
            <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
              <BookOpen className="w-10 h-10 opacity-30" />
              <p>Chưa có khóa học nào.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-bg-secondary">
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Mã</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Tên khóa học</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Loại</th>
                    <th className="text-right px-4 py-3 text-text-secondary font-medium">CPE (giờ)</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Nhà cung cấp</th>
                    <th className="text-left px-4 py-3 text-text-secondary font-medium">Trạng thái</th>
                    {canWrite && <th className="px-4 py-3" />}
                  </tr>
                </thead>
                <tbody>
                  {courses.map(c => (
                    <tr key={c.id} className="border-b hover:bg-bg-secondary/50 transition-colors">
                      <td className="px-4 py-3 font-mono text-xs text-text-secondary">{c.code}</td>
                      <td className="px-4 py-3 font-medium text-text-primary">{c.name}</td>
                      <td className="px-4 py-3">
                        <Badge variant="outline">{COURSE_TYPE_LABELS[c.course_type]}</Badge>
                      </td>
                      <td className="px-4 py-3 text-right tabular-nums">{c.cpe_hours}</td>
                      <td className="px-4 py-3 text-text-secondary">{c.provider ?? '–'}</td>
                      <td className="px-4 py-3">
                        <Badge variant={c.is_active ? 'default' : 'secondary'}>
                          {c.is_active ? 'Hoạt động' : 'Ngừng'}
                        </Badge>
                      </td>
                      {canWrite && (
                        <td className="px-4 py-3">
                          <div className="flex gap-1 justify-end">
                            <Button variant="ghost" size="icon" onClick={() => openEdit(c)}>
                              <Pencil className="w-4 h-4" />
                            </Button>
                            {canDelete && (
                              <Button
                                variant="ghost"
                                size="icon"
                                className="text-danger hover:text-danger"
                                onClick={() => setDeleteTarget(c)}
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            )}
                          </div>
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

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center gap-2 justify-end text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            Trước
          </Button>
          <span className="text-text-secondary">Trang {page} / {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            Sau
          </Button>
        </div>
      )}

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Thêm khóa học mới</DialogTitle></DialogHeader>
          <form id="course-create-form" onSubmit={submitCreate} className="flex flex-col gap-3">
            <CourseFormFields form={form} onChange={setForm} showCode />
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="course-create-form" loading={createMutation.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      <Dialog open={!!editTarget} onOpenChange={v => !v && setEditTarget(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Chỉnh sửa khóa học</DialogTitle></DialogHeader>
          <form id="course-edit-form" onSubmit={submitEdit} className="flex flex-col gap-3">
            <CourseFormFields form={form} onChange={setForm} showCode={false} />
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditTarget(null)}>Hủy</Button>
            <Button type="submit" form="course-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm dialog */}
      <Dialog open={!!deleteTarget} onOpenChange={v => !v && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Xóa khóa học?</DialogTitle></DialogHeader>
          <p className="text-sm text-text-secondary">
            Xóa khóa học <strong>{deleteTarget?.name}</strong>? Hành động này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>Hủy</Button>
            <Button
              variant="danger"
              loading={deleteMutation.isPending}
              onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
            >
              Xóa
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CourseFormFields({
  form,
  onChange,
  showCode,
}: {
  form: CourseFormState;
  onChange: (f: CourseFormState) => void;
  showCode: boolean;
}) {
  const set = (key: keyof CourseFormState) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      onChange({ ...form, [key]: e.target.value });

  return (
    <>
      {showCode && (
        <div className="flex flex-col gap-1">
          <Label>Mã khóa học <span className="text-danger">*</span></Label>
          <Input value={form.code} onChange={set('code')} required placeholder="VD: TC-CPA-001" />
        </div>
      )}
      <div className="flex flex-col gap-1">
        <Label>Tên khóa học <span className="text-danger">*</span></Label>
        <Input value={form.name} onChange={set('name')} required />
      </div>
      <div className="flex flex-col gap-1">
        <Label>Loại khóa học</Label>
        <select
          className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary"
          value={form.course_type}
          onChange={set('course_type')}
        >
          {(['TECHNICAL', 'ETHICS', 'MANAGEMENT', 'SOFT_SKILLS', 'COMPLIANCE', 'OTHER'] as CourseType[]).map(t => (
            <option key={t} value={t}>{COURSE_TYPE_LABELS[t]}</option>
          ))}
        </select>
      </div>
      <div className="flex flex-col gap-1">
        <Label>Số giờ CPE <span className="text-danger">*</span></Label>
        <Input
          type="number"
          min={0}
          step={0.5}
          value={form.cpe_hours}
          onChange={set('cpe_hours')}
          required
        />
      </div>
      <div className="flex flex-col gap-1">
        <Label>Nhà cung cấp</Label>
        <Input value={form.provider} onChange={set('provider')} placeholder="Tùy chọn" />
      </div>
      <div className="flex flex-col gap-1">
        <Label>Mô tả</Label>
        <textarea
          className="text-sm border rounded px-3 py-2 bg-bg-primary text-text-primary min-h-[60px] resize-y"
          value={form.description}
          onChange={set('description')}
        />
      </div>
      <div className="flex items-center gap-2">
        <input
          id="is-active"
          type="checkbox"
          checked={form.is_active}
          onChange={e => onChange({ ...form, is_active: e.target.checked })}
          className="w-4 h-4"
        />
        <Label htmlFor="is-active">Đang hoạt động</Label>
      </div>
    </>
  );
}
