'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { offboardingService } from '@/services/hrm/provisioning';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { MODULE_ROLES } from '@/lib/roles';
import { ArrowLeft, CheckCircle2, Circle, CheckCheck } from 'lucide-react';

const STATUS_LABELS: Record<string, string> = {
  IN_PROGRESS: 'Đang xử lý',
  COMPLETED:   'Hoàn thành',
  CANCELLED:   'Đã hủy',
};

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  IN_PROGRESS: 'outline',
  COMPLETED:   'default',
  CANCELLED:   'secondary',
};

function InfoRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value ?? '–'}</span>
    </div>
  );
}

export default function OffboardingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [noteMap, setNoteMap] = React.useState<Record<string, string>>({});
  const [savingKey, setSavingKey] = React.useState<string | null>(null);

  const canUpdateItems = MODULE_ROLES.hrmOffboardingItems.some(r => user?.roles?.includes(r));
  const canComplete    = MODULE_ROLES.hrmOffboardingCreate.some(r => user?.roles?.includes(r));

  const { data: checklist, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'offboarding', id],
    queryFn: () => offboardingService.getById(id),
    enabled: !!id,
  });

  function invalidate() {
    qc.invalidateQueries({ queryKey: ['hrm', 'offboarding', id] });
    qc.invalidateQueries({ queryKey: ['hrm', 'offboarding'] });
  }

  const completeMutation = useMutation({
    mutationFn: () => offboardingService.complete(id),
    onSuccess: () => { toast('Checklist đã hoàn thành', 'success'); invalidate(); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  async function toggleItem(key: string, currentCompleted: boolean) {
    if (!canUpdateItems) return;
    setSavingKey(key);
    try {
      await offboardingService.updateItem(id, key, {
        completed: !currentCompleted,
        notes: noteMap[key] ?? undefined,
      });
      invalidate();
    } catch (err) {
      toast(getErrorMessage(err), 'error');
    } finally {
      setSavingKey(null);
    }
  }

  async function saveNote(key: string, completed: boolean) {
    if (!canUpdateItems) return;
    setSavingKey(key);
    try {
      await offboardingService.updateItem(id, key, { completed, notes: noteMap[key] });
      toast('Đã lưu ghi chú', 'success');
      invalidate();
    } catch (err) {
      toast(getErrorMessage(err), 'error');
    } finally {
      setSavingKey(null);
    }
  }

  if (isLoading) return <PageSpinner />;

  if (isError || !checklist) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải checklist.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const items = checklist.items?.items ?? [];
  const allDone = items.length > 0 && items.every(i => i.completed);
  const isActive = checklist.status === 'IN_PROGRESS';
  const completedCount = items.filter(i => i.completed).length;

  return (
    <div className="flex flex-col gap-4 max-w-2xl">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/offboarding')}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-text-primary">Chi tiết Offboarding</h1>
            <p className="text-sm text-text-secondary font-mono">{checklist.id}</p>
          </div>
        </div>
        <Badge variant={STATUS_VARIANTS[checklist.status] ?? 'outline'}>
          {STATUS_LABELS[checklist.status] ?? checklist.status}
        </Badge>
      </div>

      <Card>
        <CardContent className="p-5">
          <div className="grid grid-cols-2 gap-4">
            <InfoRow label="Nhân viên ID" value={checklist.employee_id} />
            <InfoRow label="Loại" value={checklist.checklist_type === 'OFFBOARDING' ? 'Offboarding' : 'Onboarding'} />
            <InfoRow label="Ngày mục tiêu" value={checklist.target_date ? formatDate(checklist.target_date) : undefined} />
            <InfoRow label="Người tạo" value={checklist.initiated_by} />
            <InfoRow label="Ngày tạo" value={formatDate(checklist.created_at)} />
            {checklist.completed_at && (
              <InfoRow label="Hoàn thành lúc" value={formatDate(checklist.completed_at)} />
            )}
            {checklist.notes && (
              <div className="col-span-2"><InfoRow label="Ghi chú" value={checklist.notes} /></div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-5">
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm font-semibold text-text-primary">
              Checklist ({completedCount}/{items.length})
            </p>
            <div className="h-2 w-40 rounded-full bg-border overflow-hidden">
              <div
                className="h-full bg-action transition-all"
                style={{ width: items.length ? `${(completedCount / items.length) * 100}%` : '0%' }}
              />
            </div>
          </div>

          {items.length === 0 && (
            <p className="text-sm text-text-secondary text-center py-8">Chưa có mục nào trong checklist.</p>
          )}

          <ul className="flex flex-col gap-3">
            {items.map((item) => (
              <li key={item.key} className="flex flex-col gap-1.5">
                <div className="flex items-start gap-3">
                  <button
                    type="button"
                    disabled={!isActive || !canUpdateItems || savingKey === item.key}
                    onClick={() => toggleItem(item.key, item.completed)}
                    className="mt-0.5 flex-shrink-0 text-action disabled:opacity-40 disabled:cursor-not-allowed"
                  >
                    {item.completed
                      ? <CheckCircle2 className="w-5 h-5" />
                      : <Circle className="w-5 h-5 text-text-secondary" />
                    }
                  </button>
                  <div className="flex-1">
                    <p className={`text-sm ${item.completed ? 'line-through text-text-secondary' : 'text-text-primary'}`}>
                      {item.label}
                    </p>
                    {isActive && canUpdateItems && (
                      <div className="flex items-center gap-2 mt-1">
                        <input
                          type="text"
                          value={noteMap[item.key] ?? item.notes ?? ''}
                          onChange={e => setNoteMap(prev => ({ ...prev, [item.key]: e.target.value }))}
                          placeholder="Ghi chú..."
                          className="h-7 flex-1 rounded border border-border bg-surface-paper px-2 text-xs text-text-primary focus:outline-none focus:ring-1 focus:ring-action/30"
                        />
                        {noteMap[item.key] !== undefined && noteMap[item.key] !== (item.notes ?? '') && (
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-7 px-2 text-xs"
                            loading={savingKey === item.key}
                            onClick={() => saveNote(item.key, item.completed)}
                          >
                            Lưu
                          </Button>
                        )}
                      </div>
                    )}
                    {(!isActive || !canUpdateItems) && item.notes && (
                      <p className="text-xs text-text-secondary mt-0.5">{item.notes}</p>
                    )}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      {isActive && canComplete && (
        <div className="flex justify-end">
          <Button
            disabled={!allDone}
            loading={completeMutation.isPending}
            onClick={() => completeMutation.mutate()}
            title={!allDone ? 'Hoàn thành tất cả các mục trước' : undefined}
          >
            <CheckCheck className="w-4 h-4" />Đánh dấu hoàn thành
          </Button>
        </div>
      )}
    </div>
  );
}
