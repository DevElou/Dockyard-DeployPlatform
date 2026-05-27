"use client";

import { useState } from "react";
import { Plus, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { RuntimeTargetRow } from "@/components/runtime-targets/runtime-target-row";
import { AddTargetDialog } from "@/components/runtime-targets/add-target-dialog";
import { useRuntimeTargets } from "@/lib/hooks/use-runtime-targets";

export default function RuntimeTargetsPage() {
  const [open, setOpen] = useState(false);
  const query = useRuntimeTargets();

  return (
    <div>
      <PageHeader
        title="Runtime targets"
        description="Manage Docker hosts where deployments run"
      >
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          Add target
        </Button>
      </PageHeader>

      <DataGuard
        {...query}
        errorMessage="Could not load runtime targets."
        onRetry={query.refetch}
        loadingRows={3}
        empty={
          <EmptyState
            icon={Server}
            title="No runtime targets"
            description="Add a Docker host with a running deploy-agent to get started."
          >
            <Button size="sm" onClick={() => setOpen(true)}>
              <Plus className="mr-1.5 h-4 w-4" />
              Add target
            </Button>
          </EmptyState>
        }
      >
        {(data) => (
          <div className="p-6 space-y-3">
            {data.items.map((target) => (
              <RuntimeTargetRow key={target.id} target={target} />
            ))}
          </div>
        )}
      </DataGuard>

      <AddTargetDialog open={open} onOpenChange={setOpen} />
    </div>
  );
}
