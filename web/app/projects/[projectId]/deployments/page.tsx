"use client";

import { use, useState } from "react";
import { Rocket, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { DeploymentsTable } from "@/components/projects/deployments-table";
import { NewDeploymentDialog } from "@/components/projects/new-deployment-dialog";
import { DeploymentDetailsSheet } from "@/components/projects/deployment-details-sheet";
import { useDeployments } from "@/lib/hooks/use-deployments";
import type { Deployment } from "@/lib/types/deployment";

export default function DeploymentsPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<Deployment | null>(null);
  const query = useDeployments(projectId);

  const liveSelected =
    selected && query.data
      ? (query.data.items.find((d) => d.id === selected.id) ?? selected)
      : selected;

  return (
    <div>
      <PageHeader title="Deployments">
        <Button variant="outline" size="sm" onClick={() => query.refetch()}>
          <RefreshCw className="mr-1.5 h-4 w-4" />
          Refresh
        </Button>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Rocket className="mr-1.5 h-4 w-4" />
          Deploy
        </Button>
      </PageHeader>

      <DataGuard
        {...query}
        errorMessage="Could not load deployments."
        onRetry={query.refetch}
        empty={
          <EmptyState
            icon={Rocket}
            title="No deployments yet"
            description="Choose a release and runtime target to deploy."
          >
            <Button size="sm" onClick={() => setOpen(true)}>
              <Rocket className="mr-1.5 h-4 w-4" />
              Deploy
            </Button>
          </EmptyState>
        }
      >
        {(data) => (
          <DeploymentsTable items={data.items} onSelect={setSelected} />
        )}
      </DataGuard>

      <NewDeploymentDialog projectId={projectId} open={open} onOpenChange={setOpen} />

      <DeploymentDetailsSheet
        projectId={projectId}
        deployment={liveSelected}
        open={selected !== null}
        onOpenChange={(o) => {
          if (!o) setSelected(null);
        }}
      />
    </div>
  );
}
