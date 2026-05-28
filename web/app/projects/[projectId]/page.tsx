"use client";

import { use } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { BuildStatusBadge } from "@/components/status/build-status-badge";
import { DeploymentStatusBadge } from "@/components/status/deployment-status-badge";
import { useReleases } from "@/lib/hooks/use-releases";
import { useDeployments } from "@/lib/hooks/use-deployments";
import { formatRelative } from "@/lib/format";
import { Rocket, PackageOpen } from "lucide-react";

export default function ProjectOverviewPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const releases = useReleases(projectId);
  const deployments = useDeployments(projectId);

  const latestRelease = releases.data?.items[0];
  const latestDeployment = deployments.data?.items[0];

  if (releases.isLoading || deployments.isLoading) {
    return <LoadingState rows={3} />;
  }

  if (releases.isError || deployments.isError) {
    return <ErrorState message="Could not load project data." />;
  }

  if (!latestRelease && !latestDeployment) {
    return (
      <EmptyState
        icon={PackageOpen}
        title="No activity yet"
        description="Create a release to build and deploy this project."
      >
        <Button asChild size="sm">
          <Link href={`/projects/${projectId}/releases`}>Go to Releases</Link>
        </Button>
      </EmptyState>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {latestDeployment && (
        <section>
          <h2 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wide">
            Latest deployment
          </h2>
          <div className="rounded-lg border p-4 flex items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <DeploymentStatusBadge status={latestDeployment.status} />
                <span className="text-sm text-muted-foreground">
                  {formatRelative(latestDeployment.createdAt)}
                </span>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                Release: {latestDeployment.releaseId.slice(0, 8)} · Strategy:{" "}
                {latestDeployment.strategy}
              </p>
            </div>
            <Button variant="outline" size="sm" asChild>
              <Link href={`/projects/${projectId}/deployments`}>
                View all
              </Link>
            </Button>
          </div>
        </section>
      )}

      {latestRelease && (
        <section>
          <h2 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wide">
            Latest release
          </h2>
          <div className="rounded-lg border p-4 flex items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <BuildStatusBadge status={latestRelease.buildStatus} />
                <span className="text-sm font-mono">{latestRelease.version}</span>
                <span className="text-sm text-muted-foreground">
                  {formatRelative(latestRelease.createdAt)}
                </span>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                ref: {latestRelease.gitRef}
                {latestRelease.gitSha && (
                  <> · {latestRelease.gitSha.slice(0, 7)}</>
                )}
              </p>
            </div>
            <div className="flex gap-2">
              {latestRelease.buildStatus === "succeeded" && !latestDeployment && (
                <Button size="sm" asChild>
                  <Link href={`/projects/${projectId}/deployments`}>
                    <Rocket className="mr-1.5 h-4 w-4" />
                    Deploy
                  </Link>
                </Button>
              )}
              <Button variant="outline" size="sm" asChild>
                <Link href={`/projects/${projectId}/releases`}>View all</Link>
              </Button>
            </div>
          </div>
        </section>
      )}
    </div>
  );
}
