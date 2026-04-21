'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { PageSpinner } from '@/components/ui/spinner';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { MatrixGrid } from '@/components/hrm/matrix-grid';
import { branchService, departmentService, matrixService } from '@/services/hrm/organization';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { Plus } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO'];

export default function MatrixPage() {
  const qc = useQueryClient();
  const { user } = useAuthStore();
  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));

  const [linkOpen, setLinkOpen] = React.useState(false);
  const [linkBranch, setLinkBranch] = React.useState('');
  const [linkDept, setLinkDept] = React.useState('');

  const { data: branchData, isLoading: bLoading } = useQuery({
    queryKey: ['hrm', 'branches', 'all'],
    queryFn: () => branchService.list({ size: 100 }),
  });

  const { data: deptData, isLoading: dLoading } = useQuery({
    queryKey: ['hrm', 'departments', 'all'],
    queryFn: () => departmentService.list({ size: 100 }),
  });

  const { data: matrixData, isLoading: mLoading } = useQuery({
    queryKey: ['hrm', 'matrix'],
    queryFn: () => matrixService.list(),
  });

  const linkMutation = useMutation({
    mutationFn: ({ branchId, deptId }: { branchId: string; deptId: string }) =>
      matrixService.link(branchId, deptId),
    onSuccess: () => {
      toast('Đã liên kết phòng ban với chi nhánh', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'matrix'] });
      setLinkOpen(false);
      setLinkBranch('');
      setLinkDept('');
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const unlinkMutation = useMutation({
    mutationFn: ({ branchId, deptId }: { branchId: string; deptId: string }) =>
      matrixService.unlink(branchId, deptId),
    onSuccess: () => {
      toast('Đã hủy liên kết', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'matrix'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const isLoading = bLoading || dLoading || mLoading;

  if (isLoading) return <PageSpinner />;

  const branches = branchData?.data ?? [];
  const departments = deptData?.data ?? [];
  const links = matrixData?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Ma trận Chi nhánh × Phòng ban</h1>
          <p className="text-sm text-text-secondary">Quản lý mối quan hệ giữa chi nhánh và phòng ban</p>
        </div>
        {canWrite && (
          <Button onClick={() => setLinkOpen(true)}>
            <Plus className="w-4 h-4" />Thêm liên kết
          </Button>
        )}
      </div>

      <Card>
        <CardContent className="p-4">
          {branches.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-12 text-text-secondary">
              <p>Chưa có dữ liệu chi nhánh hoặc phòng ban để hiển thị.</p>
            </div>
          ) : (
            <>
              <div className="flex items-center gap-4 mb-3 text-xs text-text-secondary">
                <div className="flex items-center gap-1.5">
                  <span className="w-4 h-4 flex items-center justify-center rounded bg-success/10 text-success font-bold">✓</span>
                  <span>Đã liên kết</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-4 h-4 flex items-center justify-center rounded bg-background border border-border text-text-secondary">–</span>
                  <span>Chưa liên kết</span>
                </div>
                {canWrite && <span className="text-text-secondary/70 italic">Nhấp ô để thay đổi trạng thái</span>}
              </div>
              <MatrixGrid
                branches={branches}
                departments={departments}
                links={links}
                canWrite={canWrite}
                onLink={(branchId, deptId) => linkMutation.mutate({ branchId, deptId })}
                onUnlink={(branchId, deptId) => {
                  if (confirm('Hủy liên kết phòng ban này?')) {
                    unlinkMutation.mutate({ branchId, deptId });
                  }
                }}
              />
            </>
          )}
        </CardContent>
      </Card>

      {/* Link dialog */}
      <Dialog open={linkOpen} onOpenChange={setLinkOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>Thêm liên kết Chi nhánh — Phòng ban</DialogTitle></DialogHeader>
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Chi nhánh</Label>
              <Select value={linkBranch} onValueChange={setLinkBranch}>
                <SelectTrigger>
                  <SelectValue placeholder="Chọn chi nhánh..." />
                </SelectTrigger>
                <SelectContent>
                  {branches.map(b => (
                    <SelectItem key={b.id} value={b.id}>{b.name} ({b.code})</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <Label>Phòng ban</Label>
              <Select value={linkDept} onValueChange={setLinkDept}>
                <SelectTrigger>
                  <SelectValue placeholder="Chọn phòng ban..." />
                </SelectTrigger>
                <SelectContent>
                  {departments.map(d => (
                    <SelectItem key={d.id} value={d.id}>{d.name} ({d.code})</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setLinkOpen(false)}>Hủy</Button>
            <Button
              disabled={!linkBranch || !linkDept}
              loading={linkMutation.isPending}
              onClick={() => linkMutation.mutate({ branchId: linkBranch, deptId: linkDept })}
            >
              Liên kết
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
