"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { use } from "react";
import { cn } from "@/lib/utils";
import { useProject } from "@/lib/hooks/use-projects";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { ChevronRight } from "lucide-react";

const tabs = [
  { label: "Overview", suffix: "" },
  { label: "Releases", suffix: "/releases" },
  { label: "Deployments", suffix: "/deployments" },
  { label: "Services", suffix: "/services" },
];

export default function ProjectLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const pathname = usePathname();
  const { data: project, isLoading } = useProject(projectId);

  return (
    <div className="flex flex-col h-full">
      <div className="border-b px-6 pt-4 pb-0">
        <div className="flex items-center gap-1.5 text-sm text-muted-foreground mb-1">
          <Link href="/projects" className="hover:text-foreground transition-colors">
            Projects
          </Link>
          <ChevronRight className="h-3.5 w-3.5" />
          {isLoading ? (
            <Skeleton className="h-4 w-24" />
          ) : (
            <span className="text-foreground font-medium">{project?.name}</span>
          )}
        </div>

        {project && (
          <p className="text-xs text-muted-foreground mb-3">
            {project.githubOwner}/{project.githubRepo} · {project.defaultBranch}
            <Badge variant="secondary" className="ml-2 capitalize text-xs">
              {project.status}
            </Badge>
          </p>
        )}

        <nav className="flex gap-1 -mb-px">
          {tabs.map(({ label, suffix }) => {
            const href = `/projects/${projectId}${suffix}`;
            const active =
              suffix === ""
                ? pathname === `/projects/${projectId}`
                : pathname.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "px-3 py-2 text-sm border-b-2 transition-colors",
                  active
                    ? "border-primary text-foreground font-medium"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                )}
              >
                {label}
              </Link>
            );
          })}
        </nav>
      </div>

      <div className="flex-1 overflow-y-auto">{children}</div>
    </div>
  );
}
