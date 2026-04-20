'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { PageSpinner } from '@/components/ui/spinner';
import { commissionService, type PlanType, type TriggerOn } from '@/services/commissions';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatCurrency } from '@/lib/utils';
import { Plus, ShieldCheck, XCircle } from 'lucide-react';

const PLAN_TYPE_LABELS: Record<PlanType, string> = {
  flat: 'Tỷ lệ cố định', tiered: 'Bậc thang', fixed: 'Số tiền cố định', custom: 'Tùy chỉnh',
};
const TRIGGER_LABELS: Record<TriggerOn, string> = {
  invoice_issued: 'Khi xuất hóa đơn',
  payment_received: 'Khi nhận thanh toán',
  engagement_completed: 'Khi hoàn thành hợp đồng',
};

function CommissionPlansSection() {
  const qc = useQueryClient();
  const [showForm, setShowForm] = React.useState(false);
  const [form, setForm] = React.useState({
    name: '', type: 'flat' as PlanType,
    trigger_on: 'payment_received' as TriggerOn,
    base_rate: '', max_amount: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['commission-plans'],
    queryFn: () => commissionService.plans.list({ size: 50 }),
  });

  const createMut = useMutation({
    mutationFn: () => commissionService.plans.create({
      name: form.name,
      type: form.type,
      trigger_on: form.trigger_on,
      base_rate: form.base_rate ? parseFloat(form.base_rate) : undefined,
      max_amount: form.max_amount ? parseFloat(form.max_amount) : undefined,
    }),
    onSuccess: () => {
      toast('Tạo plan thành công', 'success');
      qc.invalidateQueries({ queryKey: ['commission-plans'] });
      setShowForm(false);
      setForm({ name: '', type: 'flat', trigger_on: 'payment_received', base_rate: '', max_amount: '' });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deactivateMut = useMutation({
    mutationFn: (id: string) => commissionService.plans.deactivate(id),
    onSuccess: () => { toast('Đã vô hiệu hoá plan', 'success'); qc.invalidateQueries({ queryKey: ['commission-plans'] }); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Commission Plans</CardTitle>
          <Button size="sm" onClick={() => setShowForm(v => !v)}>
            <Plus className="w-4 h-4" />{showForm ? 'Huỷ' : 'Tạo plan'}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {showForm && (
          <div className="border border-border rounded-card p-4 flex flex-col gap-3 bg-surface">
            <Input
              placeholder="Tên plan *"
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
            />
            <div className="grid grid-cols-2 gap-3">
              <Select value={form.type} onValueChange={v => setForm(f => ({ ...f, type: v as PlanType }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(PLAN_TYPE_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                </SelectContent>
              </Select>
              <Select value={form.trigger_on} onValueChange={v => setForm(f => ({ ...f, trigger_on: v as TriggerOn }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(TRIGGER_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Input
                placeholder="Tỷ lệ (%) — vd: 5"
                value={form.base_rate}
                onChange={e => setForm(f => ({ ...f, base_rate: e.target.value }))}
                type="number" min={0} max={100}
              />
              <Input
                placeholder="Hoa hồng tối đa (VNĐ)"
                value={form.max_amount}
                onChange={e => setForm(f => ({ ...f, max_amount: e.target.value }))}
                type="number" min={0}
              />
            </div>
            <Button
              size="sm"
              disabled={!form.name || createMut.isPending}
              onClick={() => createMut.mutate()}
            >
              {createMut.isPending ? 'Đang tạo...' : 'Tạo plan'}
            </Button>
          </div>
        )}

        {isLoading ? <PageSpinner /> : (
          <div className="flex flex-col gap-2">
            {data?.data.map(plan => (
              <div key={plan.id} className="flex items-center justify-between py-3 border-b border-border last:border-0">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-text-primary">{plan.name}</span>
                    <Badge variant={plan.is_active ? 'success' : 'ghost'}>
                      {plan.is_active ? 'Đang dùng' : 'Vô hiệu'}
                    </Badge>
                  </div>
                  <p className="text-xs text-text-secondary mt-0.5">
                    {PLAN_TYPE_LABELS[plan.type]} · {TRIGGER_LABELS[plan.trigger_on]}
                    {plan.base_rate != null ? ` · ${plan.base_rate}%` : ''}
                    {plan.max_amount ? ` · tối đa ${formatCurrency(plan.max_amount)}` : ''}
                  </p>
                </div>
                {plan.is_active && (
                  <Button
                    variant="ghost" size="icon"
                    title="Vô hiệu hoá"
                    onClick={() => deactivateMut.mutate(plan.id)}
                  >
                    <XCircle className="w-4 h-4 text-danger" />
                  </Button>
                )}
              </div>
            ))}
            {data?.data.length === 0 && (
              <p className="text-sm text-text-secondary text-center py-6">Chưa có commission plan nào</p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function SecuritySection() {
  const { user } = useAuthStore();

  return (
    <Card>
      <CardHeader><CardTitle>Bảo mật</CardTitle></CardHeader>
      <CardContent className="flex flex-col gap-3">
        <div className="flex items-center justify-between py-2.5 border-b border-border">
          <div>
            <p className="text-sm font-medium text-text-primary">Xác thực hai yếu tố (2FA)</p>
            <p className="text-xs text-text-secondary mt-0.5">Bảo vệ tài khoản bằng mã xác thực bổ sung</p>
          </div>
          <Badge variant={user?.two_factor_enabled ? 'success' : 'warning'}>
            <ShieldCheck className="w-3 h-3 mr-1" />
            {user?.two_factor_enabled ? 'Đã bật' : 'Chưa bật'}
          </Badge>
        </div>
        <div className="flex items-center justify-between py-2.5">
          <div>
            <p className="text-sm font-medium text-text-primary">Email</p>
            <p className="text-xs text-text-secondary mt-0.5">{user?.email}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default function SettingsPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">Cài đặt hệ thống</h2>
        <p className="text-sm text-text-secondary">Quản lý commission plans và cấu hình bảo mật</p>
      </div>
      <SecuritySection />
      <CommissionPlansSection />
    </div>
  );
}
