"use client";

import { RefreshCw } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { BuildStatusBadge } from "@/components/status/build-status-badge";
import { EventTimeline } from "@/components/observability/event-timeline";
import { useReleaseEvents } from "@/lib/hooks/use-release-events";
import { formatDate } from "@/lib/format";
import type { Release } from "@/lib/types/release";

interface ReleaseDetailsSheetProps {
  projectId: string;
  release: Release | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ReleaseDetailsSheet({
  projectId,
  release,
  open,
  onOpenChange,
}: ReleaseDetailsSheetProps) {
  const events = useReleaseEvents(projectId, release?.id ?? "", {
    enabled: open && Boolean(release),
    buildStatus: release?.buildStatus,
  });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-xl overflow-y-auto">
        {release ? (
          <>
            <SheetHeader>
              <div className="flex items-center gap-2">
                <SheetTitle className="font-mono text-base">
                  {release.version}
                </SheetTitle>
                <BuildStatusBadge status={release.buildStatus} />
              </div>
              <SheetDescription className="font-mono text-xs">
                {release.gitRef}
                {release.gitSha && <> · {release.gitSha.slice(0, 7)}</>}
              </SheetDescription>
            </SheetHeader>

            <div className="mt-4 space-y-1 text-xs text-muted-foreground">
              <div>Created {formatDate(release.createdAt)}</div>
              {release.imageDigest && (
                <div className="font-mono break-all">
                  digest: {release.imageDigest}
                </div>
              )}
            </div>

            <div className="mt-6">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Timeline
                </h3>
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
              </div>

              {events.isError ? (
                <p className="text-sm text-destructive">
                  Could not load events.
                </p>
              ) : events.isLoading ? (
                <p className="text-sm text-muted-foreground">Loading events…</p>
              ) : (
                <EventTimeline events={events.data?.items ?? []} />
              )}
            </div>
          </>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}
