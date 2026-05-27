"use client";

import { use, useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { ServiceCard } from "@/components/projects/service-card";
import { AddServiceDialog } from "@/components/projects/add-service-dialog";
import { useServices } from "@/lib/hooks/use-services";

export default function ServicesPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [addOpen, setAddOpen] = useState(false);
  const query = useServices(projectId);

  return (
    <div>
      <PageHeader
        title="Services"
        description="Manage services, environment variables, and domains"
      >
        <Button size="sm" onClick={() => setAddOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          Add service
        </Button>
      </PageHeader>

      <DataGuard
        {...query}
        errorMessage="Could not load services."
        onRetry={query.refetch}
        loadingRows={3}
        empty={
          <EmptyState
            title="No services"
            description="Add a service to manage its environment variables and domains."
          >
            <Button size="sm" onClick={() => setAddOpen(true)}>
              <Plus className="mr-1.5 h-4 w-4" />
              Add service
            </Button>
          </EmptyState>
        }
      >
        {(data) => (
          <div className="p-6 space-y-6">
            {data.items.map((service) => (
              <ServiceCard key={service.id} projectId={projectId} service={service} />
            ))}
          </div>
        )}
      </DataGuard>

      <AddServiceDialog projectId={projectId} open={addOpen} onOpenChange={setAddOpen} />
    </div>
  );
}
