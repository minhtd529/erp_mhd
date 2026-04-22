'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { profileService, type UpdateProfileRequest } from '@/services/hrm/employee';
import {
  certificationService,
  cpeRequirementService,
  type Certification,
  type CPESummary,
  type CertType,
  type CertStatus,
  type CourseType,
} from '@/services/hrm/professional';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { User, Pencil, Award, BarChart3 } from 'lucide-react';

const profileSchema = z.object({
  display_name: z.string().optional(),
  personal_phone: z.string().optional(),
  personal_email: z.string().email('Email không hợp lệ').optional().or(z.literal('')),
  current_address: z.string().optional(),
  permanent_address: z.string().optional(),
});
type ProfileFormData = z.infer<typeof profileSchema>;

type TabKey = 'info' | 'certifications' | 'cpe';

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

const CURRENT_YEAR = new Date().getFullYear();

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

// ─── Certifications tab (read-only for self) ───────────────────────────────────

function MyCertificationsTab({ employeeId }: { employeeId: string }) {
  const { data, isLoading } = useQuery({
    queryKey: ['hrm', 'employees', employeeId, 'certifications'],
    queryFn: () => certificationService.listByEmployee(employeeId, { size: 50 }),
    enabled: !!employeeId,
  });

  const certs: Certification[] = data?.data ?? [];

  if (isLoading) return <PageSpinner />;

  if (certs.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
        <Award className="w-8 h-8 opacity-30" />
        <p className="text-sm">Chưa có chứng chỉ nào.</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-bg-secondary">
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Loại</th>
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Số hiệu</th>
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Cơ quan cấp</th>
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Ngày cấp</th>
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Hết hạn</th>
            <th className="text-left px-4 py-3 text-text-secondary font-medium">Trạng thái</th>
          </tr>
        </thead>
        <tbody>
          {certs.map(cert => (
            <tr key={cert.id} className="border-b hover:bg-bg-secondary/50">
              <td className="px-4 py-3">
                <Badge variant="outline">{CERT_TYPE_LABELS[cert.cert_type]}</Badge>
              </td>
              <td className="px-4 py-3 font-mono text-xs">{cert.cert_number}</td>
              <td className="px-4 py-3 text-text-secondary">{cert.issuing_authority}</td>
              <td className="px-4 py-3">{formatDate(cert.issued_date)}</td>
              <td className="px-4 py-3">{cert.expiry_date ? formatDate(cert.expiry_date) : '–'}</td>
              <td className="px-4 py-3">
                <Badge variant={CERT_STATUS_VARIANTS[cert.status]}>
                  {CERT_STATUS_LABELS[cert.status]}
                </Badge>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ─── CPE summary tab ───────────────────────────────────────────────────────────

function CPESummaryTab({ employeeId }: { employeeId: string }) {
  const [year, setYear] = React.useState(CURRENT_YEAR);

  const { data: summary, isLoading } = useQuery({
    queryKey: ['hrm', 'employees', employeeId, 'cpe-summary', year],
    queryFn: () => cpeRequirementService.getSummary(employeeId, year),
    enabled: !!employeeId,
  });

  const yearOptions = Array.from({ length: 5 }, (_, i) => CURRENT_YEAR - 2 + i);

  const pct = summary
    ? summary.required_hours > 0
      ? Math.min(100, Math.round((summary.total_hours_earned / summary.required_hours) * 100))
      : 100
    : 0;

  return (
    <div className="flex flex-col gap-4">
      {/* Year selector */}
      <div className="flex items-center gap-3">
        <Label>Năm</Label>
        <select
          className="text-sm border rounded px-3 py-1.5 bg-bg-primary text-text-primary"
          value={year}
          onChange={e => setYear(Number(e.target.value))}
        >
          {yearOptions.map(y => <option key={y} value={y}>{y}</option>)}
        </select>
      </div>

      {isLoading ? (
        <PageSpinner />
      ) : !summary ? (
        <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
          <BarChart3 className="w-8 h-8 opacity-30" />
          <p className="text-sm">Không có dữ liệu CPE cho năm {year}.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {/* Summary cards */}
          <div className="grid grid-cols-3 gap-3">
            <Card>
              <CardContent className="p-4 text-center">
                <p className="text-2xl font-bold text-text-primary tabular-nums">
                  {summary.total_hours_earned}
                </p>
                <p className="text-xs text-text-secondary mt-1">Giờ đã hoàn thành</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <p className="text-2xl font-bold text-text-primary tabular-nums">
                  {summary.required_hours}
                </p>
                <p className="text-xs text-text-secondary mt-1">Giờ yêu cầu</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <p className={`text-2xl font-bold tabular-nums ${pct >= 100 ? 'text-success' : 'text-warning'}`}>
                  {pct}%
                </p>
                <p className="text-xs text-text-secondary mt-1">Tiến độ</p>
              </CardContent>
            </Card>
          </div>

          {/* Progress bar */}
          <div>
            <div className="flex justify-between text-xs text-text-secondary mb-1">
              <span>Tiến độ CPE năm {year}</span>
              <span>{summary.total_hours_earned} / {summary.required_hours} giờ</span>
            </div>
            <div className="w-full h-2.5 rounded-full bg-bg-secondary overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${pct >= 100 ? 'bg-success' : 'bg-action'}`}
                style={{ width: `${pct}%` }}
              />
            </div>
          </div>

          {/* By category breakdown */}
          {Object.keys(summary.by_category).length > 0 && (
            <div>
              <p className="text-sm font-medium text-text-primary mb-2">Theo danh mục</p>
              <div className="flex flex-col gap-1.5">
                {Object.entries(summary.by_category).map(([cat, hours]) => (
                  <div key={cat} className="flex items-center justify-between text-sm">
                    <span className="text-text-secondary">
                      {COURSE_TYPE_LABELS[cat as CourseType] ?? cat}
                    </span>
                    <span className="tabular-nums font-medium">{hours} giờ</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ─── Page ──────────────────────────────────────────────────────────────────────

export default function MyProfilePage() {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [editOpen, setEditOpen] = React.useState(false);
  const [tab, setTab] = React.useState<TabKey>('info');

  const { data: profile, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'my-profile'],
    queryFn: () => profileService.get(),
  });

  const { register, handleSubmit, reset, formState: { errors } } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
  });

  React.useEffect(() => {
    if (editOpen && profile) {
      reset({
        display_name: profile.display_name ?? '',
        personal_phone: profile.personal_phone ?? '',
        personal_email: profile.personal_email ?? '',
        current_address: profile.current_address ?? '',
        permanent_address: profile.permanent_address ?? '',
      });
    }
  }, [editOpen, profile, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: UpdateProfileRequest) => profileService.update(data),
    onSuccess: () => {
      toast('Cập nhật hồ sơ thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'my-profile'] });
      setEditOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const submit = (data: ProfileFormData) => {
    updateMutation.mutate({
      display_name: data.display_name || undefined,
      personal_phone: data.personal_phone || undefined,
      personal_email: data.personal_email || undefined,
      current_address: data.current_address || undefined,
      permanent_address: data.permanent_address || undefined,
    });
  };

  if (isLoading) return <PageSpinner />;

  if (isError || !profile) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <User className="w-12 h-12 opacity-30" />
        <p>Chưa có hồ sơ nhân viên. Liên hệ HR để tạo hồ sơ.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 max-w-3xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Hồ sơ của tôi</h1>
          <p className="text-sm text-text-secondary">{user?.email}</p>
        </div>
        {tab === 'info' && (
          <Button variant="outline" onClick={() => setEditOpen(true)}>
            <Pencil className="w-4 h-4" />Chỉnh sửa
          </Button>
        )}
      </div>

      {/* Avatar + name strip */}
      <div className="flex items-center gap-3 px-1">
        <div className="w-12 h-12 rounded-full bg-action/10 flex items-center justify-center text-action font-semibold text-lg flex-shrink-0">
          {profile.full_name.charAt(0).toUpperCase()}
        </div>
        <div>
          <p className="font-semibold text-text-primary">{profile.full_name}</p>
          <div className="flex gap-1 flex-wrap">
            <Badge variant="outline">{profile.grade}</Badge>
            <Badge variant={profile.status === 'ACTIVE' ? 'default' : 'secondary'}>
              {profile.status === 'ACTIVE' ? 'Đang làm' : profile.status}
            </Badge>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b flex gap-0">
        <TabButton active={tab === 'info'} onClick={() => setTab('info')}>Thông tin</TabButton>
        <TabButton active={tab === 'certifications'} onClick={() => setTab('certifications')}>Chứng chỉ</TabButton>
        <TabButton active={tab === 'cpe'} onClick={() => setTab('cpe')}>CPE</TabButton>
      </div>

      {/* Tab: Thông tin */}
      {tab === 'info' && (
        <div className="flex flex-col gap-4">
          <Card>
            <CardContent className="p-5">
              <div className="grid grid-cols-2 gap-4">
                <InfoRow label="Mã nhân viên" value={profile.employee_code} />
                <InfoRow label="Tên hiển thị" value={profile.display_name} />
                <InfoRow label="Email công ty" value={profile.email} />
                <InfoRow label="Email cá nhân" value={profile.personal_email} />
                <InfoRow label="ĐT cá nhân" value={profile.personal_phone} />
                <InfoRow label="ĐT công ty" value={profile.work_phone} />
                <InfoRow label="Chức vụ" value={profile.position_title} />
                <InfoRow label="Địa điểm làm việc" value={profile.work_location} />
                <InfoRow label="Ngày bắt đầu" value={profile.hired_date} />
                <InfoRow label="Trình độ học vấn" value={profile.education_level} />
                <div className="col-span-2">
                  <InfoRow label="Địa chỉ hiện tại" value={profile.current_address} />
                </div>
                <div className="col-span-2">
                  <InfoRow label="Địa chỉ thường trú" value={profile.permanent_address} />
                </div>
              </div>
            </CardContent>
          </Card>

          {(profile.vn_cpa_number || profile.practicing_certificate_number || profile.education_school) && (
            <Card>
              <CardContent className="p-5">
                <p className="font-medium text-text-primary mb-3">Thông tin chuyên môn</p>
                <div className="grid grid-cols-2 gap-4">
                  <InfoRow label="Chuyên ngành" value={profile.education_major} />
                  <InfoRow label="Trường học" value={profile.education_school} />
                  <InfoRow label="Số CPA" value={profile.vn_cpa_number} />
                  <InfoRow label="Ngày cấp CPA" value={profile.vn_cpa_issued_date} />
                  <InfoRow label="Số chứng chỉ hành nghề" value={profile.practicing_certificate_number} />
                  <InfoRow label="Hết hạn chứng chỉ" value={profile.practicing_certificate_expiry} />
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Tab: Chứng chỉ */}
      {tab === 'certifications' && (
        <Card>
          <CardContent className="p-5">
            <MyCertificationsTab employeeId={profile.id} />
          </CardContent>
        </Card>
      )}

      {/* Tab: CPE */}
      {tab === 'cpe' && (
        <Card>
          <CardContent className="p-5">
            <CPESummaryTab employeeId={profile.id} />
          </CardContent>
        </Card>
      )}

      {/* Edit dialog — only self-editable fields */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Cập nhật hồ sơ cá nhân</DialogTitle>
          </DialogHeader>
          <form id="my-profile-form" onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
            <p className="text-xs text-text-secondary">
              Chỉ có thể chỉnh sửa thông tin cá nhân. Liên hệ HR để thay đổi thông tin công việc.
            </p>
            <div className="flex flex-col gap-1">
              <Label>Tên hiển thị</Label>
              <Input {...register('display_name')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>ĐT cá nhân</Label>
              <Input {...register('personal_phone')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Email cá nhân</Label>
              <Input {...register('personal_email')} type="email" />
              {errors.personal_email && <p className="text-xs text-danger">{errors.personal_email.message}</p>}
            </div>
            <div className="flex flex-col gap-1">
              <Label>Địa chỉ hiện tại</Label>
              <Input {...register('current_address')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Địa chỉ thường trú</Label>
              <Input {...register('permanent_address')} />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Hủy</Button>
            <Button type="submit" form="my-profile-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
