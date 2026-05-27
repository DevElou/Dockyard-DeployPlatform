import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import type { Project } from "@/lib/types/project";

interface ProjectCardProps {
  project: Project;
}

export function ProjectCard({ project }: ProjectCardProps) {
  return (
    <Link
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
  );
}
