"use client";

import { use, useState } from "react";
import { Plus, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { ReleasesTable } from "@/components/projects/releases-table";
import { NewReleaseDialog } from "@/components/projects/new-release-dialog";
import { useReleases } from "@/lib/hooks/use-releases";

export default function ReleasesPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [open, setOpen] = useState(false);
  const query = useReleases(projectId);

  return (
    <div>
      <PageHeader title="Releases">
        <Button variant="outline" size="sm" onClick={() => query.refetch()}>
          <RefreshCw className="mr-1.5 h-4 w-4" />
          Refresh
        </Button>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          New release
        </Button>
      </PageHeader>

      <DataGuard
        {...query}
        errorMessage="Could not load releases."
        onRetry={query.refetch}
        empty={
          <EmptyState
            title="No releases yet"
            description="Trigger a build from a Git ref to create your first release."
          >
            <Button size="sm" onClick={() => setOpen(true)}>
              <Plus className="mr-1.5 h-4 w-4" />
              New release
            </Button>
          </EmptyState>
        }
      >
        {(data) => <ReleasesTable items={data.items} />}
      </DataGuard>

      <NewReleaseDialog projectId={projectId} open={open} onOpenChange={setOpen} />
    </div>
  );
}
