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
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { User, Pencil } from 'lucide-react';

const profileSchema = z.object({
  display_name: z.string().optional(),
  personal_phone: z.string().optional(),
  personal_email: z.string().email('Email không hợp lệ').optional().or(z.literal('')),
  current_address: z.string().optional(),
  permanent_address: z.string().optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

function InfoRow({ label, value }: { label: string; value?: string | number | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value ?? '–'}</span>
    </div>
  );
}

export default function MyProfilePage() {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [editOpen, setEditOpen] = React.useState(false);

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
        <Button variant="outline" onClick={() => setEditOpen(true)}>
          <Pencil className="w-4 h-4" />Chỉnh sửa
        </Button>
      </div>

      {/* Basic info */}
      <Card>
        <CardContent className="p-5">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-full bg-action/10 flex items-center justify-center text-action font-semibold text-lg">
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

      {/* Professional info */}
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
