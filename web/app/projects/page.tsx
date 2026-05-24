"use client";

import Link from "next/link";
import { Plus, LayoutGrid } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { useProjects } from "@/lib/hooks/use-projects";

export default function ProjectsPage() {
  const { data, isLoading, isError, refetch } = useProjects();

  return (
    <div>
      <PageHeader title="Projects" description="Manage your deployment projects">
        <Button asChild size="sm">
          <Link href="/projects/new">
            <Plus className="mr-1.5 h-4 w-4" />
            New project
          </Link>
        </Button>
      </PageHeader>

      {isLoading && <LoadingState rows={4} />}
      {isError && (
        <ErrorState
          message="Could not load projects."
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && data?.items.length === 0 && (
        <EmptyState
          icon={LayoutGrid}
          title="No projects yet"
          description="Create your first project to get started."
        >
          <Button asChild size="sm">
            <Link href="/projects/new">
              <Plus className="mr-1.5 h-4 w-4" />
              New project
            </Link>
          </Button>
        </EmptyState>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <div className="p-6">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.items.map((project) => (
              <Link
                key={project.id}
                href={`/projects/${project.id}`}
                className="block rounded-lg border bg-card p-4 transition-colors hover:bg-accent/30"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="truncate font-medium">{project.name}</p>
                    <p className="mt-0.5 truncate text-xs text-muted-foreground">
                      {project.githubOwner}/{project.githubRepo}
                    </p>
                  </div>
                  <Badge
                    variant={project.status === "active" ? "outline" : "secondary"}
                    className="shrink-0 capitalize"
                  >
                    {project.status}
                  </Badge>
                </div>
                <p className="mt-3 text-xs text-muted-foreground">
                  Branch: {project.defaultBranch}
                </p>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
