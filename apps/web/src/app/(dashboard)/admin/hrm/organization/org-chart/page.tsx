'use client';
import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { PageSpinner } from '@/components/ui/spinner';
import { OrgChartTree } from '@/components/hrm/org-chart-tree';
import { orgChartService } from '@/services/hrm/organization';

export default function OrgChartPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'org-chart'],
    queryFn: () => orgChartService.get(),
  });

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải sơ đồ tổ chức.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h1 className="text-xl font-semibold text-text-primary">Sơ đồ Tổ chức</h1>
        <p className="text-sm text-text-secondary">Cấu trúc chi nhánh và phòng ban của công ty</p>
      </div>

      <Card>
        <CardContent className="p-4 min-h-64">
          <OrgChartTree branches={data?.branches ?? []} />
        </CardContent>
      </Card>
    </div>
  );
}
