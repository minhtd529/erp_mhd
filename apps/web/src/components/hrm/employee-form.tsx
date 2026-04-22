'use client';
import * as React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { EmployeeDetail, CreateEmployeeRequest, UpdateEmployeeRequest } from '@/services/hrm/employee';

// ─── Schemas ──────────────────────────────────────────────────────────────────

const createSchema = z.object({
  full_name: z.string().min(1, 'Bắt buộc').max(200),
  email: z.string().email('Email không hợp lệ').max(255),
  phone: z.string().optional(),
  date_of_birth: z.string().optional(),
  grade: z.enum(['EXECUTIVE', 'PARTNER', 'DIRECTOR', 'MANAGER', 'SENIOR', 'JUNIOR', 'INTERN', 'SUPPORT'], { required_error: 'Bắt buộc' }),
  status: z.enum(['ACTIVE', 'INACTIVE', 'ON_LEAVE', 'TERMINATED']).optional(),
  branch_id: z.string().uuid('UUID không hợp lệ').optional().or(z.literal('')),
  department_id: z.string().uuid('UUID không hợp lệ').optional().or(z.literal('')),
  position_title: z.string().optional(),
  employment_type: z.enum(['FULL_TIME', 'PART_TIME', 'INTERN']).optional(),
  hired_date: z.string().optional(),
  display_name: z.string().optional(),
  gender: z.enum(['MALE', 'FEMALE', 'OTHER']).optional(),
  personal_email: z.string().email('Email không hợp lệ').optional().or(z.literal('')),
  personal_phone: z.string().optional(),
  work_location: z.enum(['OFFICE', 'REMOTE', 'HYBRID']).optional(),
  hired_source: z.enum(['REFERRAL', 'PORTAL', 'DIRECT', 'AGENCY']).optional(),
  education_level: z.enum(['BACHELOR', 'MASTER', 'PHD', 'COLLEGE', 'OTHER']).optional(),
  commission_type: z.enum(['FIXED', 'TIERED', 'NONE']).optional(),
});

export type CreateEmployeeFormData = z.infer<typeof createSchema>;

const updateSchema = z.object({
  full_name: z.string().min(1).max(200).optional(),
  phone: z.string().optional(),
  grade: z.enum(['EXECUTIVE', 'PARTNER', 'DIRECTOR', 'MANAGER', 'SENIOR', 'JUNIOR', 'INTERN', 'SUPPORT']).optional(),
  status: z.enum(['ACTIVE', 'INACTIVE', 'ON_LEAVE', 'TERMINATED']).optional(),
  branch_id: z.string().uuid().optional().or(z.literal('')),
  department_id: z.string().uuid().optional().or(z.literal('')),
  position_title: z.string().optional(),
  employment_type: z.enum(['FULL_TIME', 'PART_TIME', 'INTERN']).optional(),
  hired_date: z.string().optional(),
  termination_date: z.string().optional(),
  termination_reason: z.string().optional(),
  display_name: z.string().optional(),
  gender: z.enum(['MALE', 'FEMALE', 'OTHER']).optional(),
  personal_email: z.string().email().optional().or(z.literal('')),
  personal_phone: z.string().optional(),
  work_phone: z.string().optional(),
  current_address: z.string().optional(),
  permanent_address: z.string().optional(),
  work_location: z.enum(['OFFICE', 'REMOTE', 'HYBRID']).optional(),
  education_level: z.enum(['BACHELOR', 'MASTER', 'PHD', 'COLLEGE', 'OTHER']).optional(),
  education_major: z.string().optional(),
  education_school: z.string().optional(),
  commission_type: z.enum(['FIXED', 'TIERED', 'NONE']).optional(),
  nationality: z.string().optional(),
});

export type UpdateEmployeeFormData = z.infer<typeof updateSchema>;

// ─── Grade / Status labels ────────────────────────────────────────────────────

const GRADE_LABELS: Record<string, string> = {
  EXECUTIVE: 'Executive', PARTNER: 'Partner', DIRECTOR: 'Director',
  MANAGER: 'Manager', SENIOR: 'Senior', JUNIOR: 'Junior', INTERN: 'Intern', SUPPORT: 'Support',
};
const STATUS_LABELS: Record<string, string> = {
  ACTIVE: 'Đang làm', INACTIVE: 'Không hoạt động', ON_LEAVE: 'Nghỉ phép', TERMINATED: 'Đã nghỉ',
};
const EMP_TYPE_LABELS: Record<string, string> = {
  FULL_TIME: 'Toàn thời gian', PART_TIME: 'Bán thời gian', INTERN: 'Thực tập',
};
const LOC_LABELS: Record<string, string> = {
  OFFICE: 'Văn phòng', REMOTE: 'Từ xa', HYBRID: 'Kết hợp',
};
const GENDER_LABELS: Record<string, string> = { MALE: 'Nam', FEMALE: 'Nữ', OTHER: 'Khác' };
const EDU_LABELS: Record<string, string> = {
  BACHELOR: 'Đại học', MASTER: 'Thạc sĩ', PHD: 'Tiến sĩ', COLLEGE: 'Cao đẳng', OTHER: 'Khác',
};

// ─── SelectField helper ───────────────────────────────────────────────────────

function SF({ label, value, onChange, options, placeholder }: {
  label: string;
  value?: string;
  onChange: (v: string) => void;
  options: Record<string, string>;
  placeholder?: string;
}) {
  return (
    <div className="flex flex-col gap-1">
      <Label>{label}</Label>
      <Select value={value ?? ''} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue placeholder={placeholder ?? 'Chọn...'} />
        </SelectTrigger>
        <SelectContent>
          {Object.entries(options).map(([k, v]) => (
            <SelectItem key={k} value={k}>{v}</SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

// ─── Create form ──────────────────────────────────────────────────────────────

interface CreateEmployeeFormProps {
  formId: string;
  onSubmit: (data: CreateEmployeeRequest) => void;
}

export function CreateEmployeeForm({ formId, onSubmit }: CreateEmployeeFormProps) {
  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<CreateEmployeeFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: { status: 'ACTIVE', employment_type: 'FULL_TIME', work_location: 'OFFICE', commission_type: 'NONE' },
  });

  const submit = (data: CreateEmployeeFormData) => {
    const req: CreateEmployeeRequest = {
      ...data,
      branch_id: data.branch_id || undefined,
      department_id: data.department_id || undefined,
      personal_email: data.personal_email || undefined,
    };
    onSubmit(req);
  };

  return (
    <form id={formId} onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Họ và tên *</Label>
          <Input {...register('full_name')} />
          {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email công ty *</Label>
          <Input {...register('email')} type="email" />
          {errors.email && <p className="text-xs text-danger">{errors.email.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <SF
          label="Cấp bậc *"
          value={watch('grade')}
          onChange={(v) => setValue('grade', v as CreateEmployeeFormData['grade'])}
          options={GRADE_LABELS}
        />
        {errors.grade && <p className="text-xs text-danger col-span-2">{errors.grade.message}</p>}
        <SF
          label="Trạng thái"
          value={watch('status')}
          onChange={(v) => setValue('status', v as CreateEmployeeFormData['status'])}
          options={STATUS_LABELS}
        />
        <SF
          label="Loại hợp đồng"
          value={watch('employment_type')}
          onChange={(v) => setValue('employment_type', v as CreateEmployeeFormData['employment_type'])}
          options={EMP_TYPE_LABELS}
        />
        <SF
          label="Địa điểm làm việc"
          value={watch('work_location')}
          onChange={(v) => setValue('work_location', v as CreateEmployeeFormData['work_location'])}
          options={LOC_LABELS}
        />
        <SF
          label="Giới tính"
          value={watch('gender')}
          onChange={(v) => setValue('gender', v as CreateEmployeeFormData['gender'])}
          options={GENDER_LABELS}
        />
        <div className="flex flex-col gap-1">
          <Label>Ngày sinh</Label>
          <Input {...register('date_of_birth')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày bắt đầu</Label>
          <Input {...register('hired_date')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chi nhánh (UUID)</Label>
          <Input {...register('branch_id')} placeholder="xxxxxxxx-xxxx-..." />
          {errors.branch_id && <p className="text-xs text-danger">{errors.branch_id.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Phòng ban (UUID)</Label>
          <Input {...register('department_id')} placeholder="xxxxxxxx-xxxx-..." />
          {errors.department_id && <p className="text-xs text-danger">{errors.department_id.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chức vụ</Label>
          <Input {...register('position_title')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Tên hiển thị</Label>
          <Input {...register('display_name')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email cá nhân</Label>
          <Input {...register('personal_email')} type="email" />
          {errors.personal_email && <p className="text-xs text-danger">{errors.personal_email.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>ĐT cá nhân</Label>
          <Input {...register('personal_phone')} />
        </div>
        <SF
          label="Trình độ học vấn"
          value={watch('education_level')}
          onChange={(v) => setValue('education_level', v as CreateEmployeeFormData['education_level'])}
          options={EDU_LABELS}
          placeholder="Chọn trình độ..."
        />
        <SF
          label="Nguồn tuyển dụng"
          value={watch('hired_source')}
          onChange={(v) => setValue('hired_source', v as CreateEmployeeFormData['hired_source'])}
          options={{ REFERRAL: 'Giới thiệu', PORTAL: 'Website', DIRECT: 'Trực tiếp', AGENCY: 'Agency' }}
          placeholder="Chọn nguồn..."
        />
      </div>
    </form>
  );
}

// ─── Update form ──────────────────────────────────────────────────────────────

interface UpdateEmployeeFormProps {
  formId: string;
  initial: EmployeeDetail;
  onSubmit: (data: UpdateEmployeeRequest) => void;
}

export function UpdateEmployeeForm({ formId, initial, onSubmit }: UpdateEmployeeFormProps) {
  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<UpdateEmployeeFormData>({
    resolver: zodResolver(updateSchema),
    defaultValues: {
      full_name: initial.full_name ?? '',
      phone: initial.phone ?? '',
      grade: initial.grade as UpdateEmployeeFormData['grade'],
      status: initial.status as UpdateEmployeeFormData['status'],
      branch_id: initial.branch_id ?? '',
      department_id: initial.department_id ?? '',
      position_title: initial.position_title ?? '',
      employment_type: initial.employment_type as UpdateEmployeeFormData['employment_type'],
      hired_date: initial.hired_date ?? '',
      termination_date: initial.termination_date ?? '',
      termination_reason: initial.termination_reason ?? '',
      display_name: initial.display_name ?? '',
      gender: initial.gender as UpdateEmployeeFormData['gender'],
      personal_email: initial.personal_email ?? '',
      personal_phone: initial.personal_phone ?? '',
      work_phone: initial.work_phone ?? '',
      current_address: initial.current_address ?? '',
      permanent_address: initial.permanent_address ?? '',
      work_location: initial.work_location as UpdateEmployeeFormData['work_location'],
      education_level: initial.education_level as UpdateEmployeeFormData['education_level'],
      education_major: initial.education_major ?? '',
      education_school: initial.education_school ?? '',
      commission_type: initial.commission_type as UpdateEmployeeFormData['commission_type'],
      nationality: initial.nationality ?? '',
    },
  });

  const submit = (data: UpdateEmployeeFormData) => {
    const req: UpdateEmployeeRequest = {
      ...data,
      branch_id: data.branch_id || undefined,
      department_id: data.department_id || undefined,
      personal_email: data.personal_email || undefined,
    };
    onSubmit(req);
  };

  return (
    <form id={formId} onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Họ và tên</Label>
          <Input {...register('full_name')} />
          {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Tên hiển thị</Label>
          <Input {...register('display_name')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <SF
          label="Cấp bậc"
          value={watch('grade')}
          onChange={(v) => setValue('grade', v as UpdateEmployeeFormData['grade'])}
          options={GRADE_LABELS}
        />
        <SF
          label="Trạng thái"
          value={watch('status')}
          onChange={(v) => setValue('status', v as UpdateEmployeeFormData['status'])}
          options={STATUS_LABELS}
        />
        <SF
          label="Loại hợp đồng"
          value={watch('employment_type')}
          onChange={(v) => setValue('employment_type', v as UpdateEmployeeFormData['employment_type'])}
          options={EMP_TYPE_LABELS}
        />
        <SF
          label="Địa điểm làm việc"
          value={watch('work_location')}
          onChange={(v) => setValue('work_location', v as UpdateEmployeeFormData['work_location'])}
          options={LOC_LABELS}
        />
        <div className="flex flex-col gap-1">
          <Label>Chi nhánh (UUID)</Label>
          <Input {...register('branch_id')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Phòng ban (UUID)</Label>
          <Input {...register('department_id')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chức vụ</Label>
          <Input {...register('position_title')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày bắt đầu</Label>
          <Input {...register('hired_date')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email cá nhân</Label>
          <Input {...register('personal_email')} type="email" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>ĐT cá nhân</Label>
          <Input {...register('personal_phone')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>ĐT công ty</Label>
          <Input {...register('work_phone')} />
        </div>
        <SF
          label="Giới tính"
          value={watch('gender')}
          onChange={(v) => setValue('gender', v as UpdateEmployeeFormData['gender'])}
          options={GENDER_LABELS}
        />
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ hiện tại</Label>
          <Input {...register('current_address')} />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ thường trú</Label>
          <Input {...register('permanent_address')} />
        </div>
        <SF
          label="Trình độ học vấn"
          value={watch('education_level')}
          onChange={(v) => setValue('education_level', v as UpdateEmployeeFormData['education_level'])}
          options={EDU_LABELS}
        />
        <div className="flex flex-col gap-1">
          <Label>Chuyên ngành</Label>
          <Input {...register('education_major')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Trường học</Label>
          <Input {...register('education_school')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Quốc tịch</Label>
          <Input {...register('nationality')} />
        </div>
        <SF
          label="Loại hoa hồng"
          value={watch('commission_type')}
          onChange={(v) => setValue('commission_type', v as UpdateEmployeeFormData['commission_type'])}
          options={{ FIXED: 'Cố định', TIERED: 'Bậc thang', NONE: 'Không' }}
        />
        <div className="flex flex-col gap-1">
          <Label>Lý do nghỉ việc</Label>
          <Input {...register('termination_reason')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày nghỉ việc</Label>
          <Input {...register('termination_date')} type="date" />
        </div>
      </div>
    </form>
  );
}
