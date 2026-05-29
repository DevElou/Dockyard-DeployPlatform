"use client";

import { RefreshCw } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { DeploymentStatusBadge } from "@/components/status/deployment-status-badge";
import { EventTimeline } from "@/components/observability/event-timeline";
import { ContainerLogsView } from "@/components/observability/container-logs-view";
import { useDeploymentEvents } from "@/lib/hooks/use-deployment-events";
import { useContainerLogs } from "@/lib/hooks/use-container-logs";
import { ApiError } from "@/lib/types/api";
import { formatDate } from "@/lib/format";
import { useState } from "react";
import type { Deployment } from "@/lib/types/deployment";

interface DeploymentDetailsSheetProps {
  projectId: string;
  deployment: Deployment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeploymentDetailsSheet({
  projectId,
  deployment,
  open,
  onOpenChange,
}: DeploymentDetailsSheetProps) {
  const [tab, setTab] = useState<"timeline" | "logs">("timeline");
  const isOpen = open && Boolean(deployment);

  const events = useDeploymentEvents(projectId, deployment?.id ?? "", {
    enabled: isOpen,
    status: deployment?.status,
  });

  const logs = useContainerLogs(projectId, deployment?.id ?? "", {
    // Only fetch logs when the user actively switches to the Logs tab.
    enabled: isOpen && tab === "logs",
    tail: 300,
  });

  const logsErrorMessage =
    logs.error instanceof ApiError
      ? logs.error.code === "agent_unavailable"
        ? "Container logs are disabled (the API was started without a deploy-agent key)."
        : logs.error.message
      : undefined;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-2xl overflow-y-auto">
        {deployment ? (
          <>
            <SheetHeader>
              <div className="flex items-center gap-2">
                <SheetTitle className="font-mono text-base">
                  {deployment.id.slice(0, 8)}
                </SheetTitle>
                <DeploymentStatusBadge status={deployment.status} />
              </div>
              <SheetDescription className="font-mono text-xs">
                release {deployment.releaseId.slice(0, 8)} · target{" "}
                {deployment.runtimeTargetId.slice(0, 8)} · {deployment.strategy}
              </SheetDescription>
            </SheetHeader>

            <div className="mt-4 space-y-1 text-xs text-muted-foreground">
              <div>Created {formatDate(deployment.createdAt)}</div>
              {deployment.startedAt && (
                <div>Started {formatDate(deployment.startedAt)}</div>
              )}
              {deployment.finishedAt && (
                <div>Finished {formatDate(deployment.finishedAt)}</div>
              )}
            </div>

            <Tabs
              value={tab}
              onValueChange={(v) => setTab(v as "timeline" | "logs")}
              className="mt-6"
            >
              <div className="flex items-center justify-between gap-2">
                <TabsList>
                  <TabsTrigger value="timeline">Timeline</TabsTrigger>
                  <TabsTrigger value="logs">Container logs</TabsTrigger>
                </TabsList>
                {tab === "timeline" && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => events.refetch()}
                    disabled={events.isFetching}
                  >
                    <RefreshCw
                      className={`mr-1.5 h-3.5 w-3.5 ${events.isFetching ? "animate-spin" : ""}`}
                    />
                    Refresh
                  </Button>
                )}
              </div>

              <TabsContent value="timeline" className="mt-4">
                {events.isError ? (
                  <p className="text-sm text-destructive">
                    Could not load events.
                  </p>
                ) : events.isLoading ? (
                  <p className="text-sm text-muted-foreground">
                    Loading events…
                  </p>
                ) : (
                  <EventTimeline events={events.data?.items ?? []} />
                )}
              </TabsContent>

              <TabsContent value="logs" className="mt-4">
                <ContainerLogsView
                  data={logs.data}
                  isLoading={logs.isLoading || logs.isFetching}
                  isError={logs.isError}
                  errorMessage={logsErrorMessage}
                  onRefresh={() => logs.refetch()}
                />
              </TabsContent>
            </Tabs>
          </>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
