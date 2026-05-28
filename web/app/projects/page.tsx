"use client";

import Link from "next/link";
import { Plus, LayoutGrid } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { DataGuard } from "@/components/common/data-guard";
import { ProjectCard } from "@/components/projects/project-card";
import { useProjects } from "@/lib/hooks/use-projects";

export default function ProjectsPage() {
  const query = useProjects();

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

      <DataGuard
        {...query}
        errorMessage="Could not load projects."
        onRetry={query.refetch}
        loadingRows={4}
        empty={
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
        }
      >
        {(data) => (
          <div className="p-6">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {data.items.map((project) => (
                <ProjectCard key={project.id} project={project} />
              ))}
            </div>
          </div>
        )}
      </DataGuard>
    </div>
  );
}
