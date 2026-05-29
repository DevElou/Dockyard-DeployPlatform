"use client";

import { use, useState } from "react";
import { Plus, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { ReleasesTable } from "@/components/projects/releases-table";
import { NewReleaseDialog } from "@/components/projects/new-release-dialog";
import { ReleaseDetailsSheet } from "@/components/projects/release-details-sheet";
import { useReleases } from "@/lib/hooks/use-releases";
import type { Release } from "@/lib/types/release";

export default function ReleasesPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<Release | null>(null);
  const query = useReleases(projectId);

  // Keep the panel in sync with refreshed list data (status/digest may change
  // while the panel is open).
  const liveSelected =
    selected && query.data
      ? (query.data.items.find((r) => r.id === selected.id) ?? selected)
      : selected;

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
        {(data) => (
          <ReleasesTable items={data.items} onSelect={setSelected} />
        )}
      </DataGuard>

      <NewReleaseDialog projectId={projectId} open={open} onOpenChange={setOpen} />

      <ReleaseDetailsSheet
        projectId={projectId}
        release={liveSelected}
        open={selected !== null}
        onOpenChange={(o) => {
          if (!o) setSelected(null);
        }}
      />
    </div>
  );
}
