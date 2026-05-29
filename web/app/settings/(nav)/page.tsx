"use client";

import Link from "next/link";
import { Puzzle, Info } from "lucide-react";
import { SectionCard } from "@/components/settings/section-card";
import { useHealth, useSystemInfo } from "@/lib/hooks/use-system";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export default function SettingsPage() {
  const health = useHealth();
  const sysInfo = useSystemInfo();

  const isHealthy = health.data === true;
  const version = sysInfo.data?.version ?? "—";

  return (
    <>
      <SectionCard
        title="Platform"
        description="Dockyard is a self-hosted deployment platform for Docker hosts on your homelab infrastructure."
      >
        <p className="text-sm text-muted-foreground">
          Manage your projects, releases, and deployments from a single control
          plane. Each Docker host runs a lightweight deploy-agent that executes
          deployment specs locally.
        </p>
      </SectionCard>

      <SectionCard title="Quick links">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Link
            href="/settings/integrations"
            className="flex items-center gap-3 rounded-lg border p-3 text-sm hover:bg-accent transition-colors"
          >
            <Puzzle className="h-4 w-4 text-muted-foreground shrink-0" />
            <div>
              <p className="font-medium">Integrations</p>
              <p className="text-xs text-muted-foreground">NPM, GitHub, DNS</p>
            </div>
          </Link>
          <Link
            href="/settings/about"
            className="flex items-center gap-3 rounded-lg border p-3 text-sm hover:bg-accent transition-colors"
          >
            <Info className="h-4 w-4 text-muted-foreground shrink-0" />
            <div>
              <p className="font-medium">About</p>
              <p className="text-xs text-muted-foreground">Version & status</p>
            </div>
          </Link>
        </div>
      </SectionCard>

      <SectionCard title="System status">
        <div className="space-y-3 text-sm">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">API</span>
            <div className="flex items-center gap-1.5">
              <span
                className={
                  health.isPending
                    ? "h-2 w-2 rounded-full bg-muted animate-pulse inline-block"
                    : isHealthy
                    ? "h-2 w-2 rounded-full bg-green-500 inline-block"
                    : "h-2 w-2 rounded-full bg-red-500 inline-block"
                }
              />
              <span className={isHealthy ? "text-green-600" : "text-red-600"}>
                {health.isPending
                  ? "Checking…"
                  : isHealthy
                  ? "Healthy"
                  : "Unreachable"}
              </span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Endpoint</span>
            <span className="font-mono text-xs">{BASE_URL}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Version</span>
            <span className="font-mono text-xs">{version}</span>
          </div>
        </div>
      </SectionCard>
    </>
  );
}
