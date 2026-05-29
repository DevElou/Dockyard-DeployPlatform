"use client";

import { SectionCard } from "@/components/settings/section-card";
import { useHealth, useSystemInfo } from "@/lib/hooks/use-system";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export default function AboutPage() {
  const health = useHealth();
  const sysInfo = useSystemInfo();

  const isHealthy = health.data === true;
  const version = sysInfo.data?.version ?? "—";

  return (
    <>
      <SectionCard title="Dockyard" description="Private deployment platform">
        <div className="space-y-3 text-sm">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Version</span>
            <span className="font-mono text-xs">{version}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">API endpoint</span>
            <span className="font-mono text-xs">{BASE_URL}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">API status</span>
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
              <span
                className={
                  isHealthy ? "text-green-600" : "text-muted-foreground"
                }
              >
                {health.isPending
                  ? "Checking…"
                  : isHealthy
                  ? "Healthy"
                  : "Unreachable"}
              </span>
            </div>
          </div>
        </div>
      </SectionCard>

      <SectionCard title="Resources">
        <div className="space-y-2 text-sm">
          <a
            href="https://github.com/elouan/dockyard"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-between py-2 border-b last:border-0 hover:text-foreground text-muted-foreground transition-colors"
          >
            <span>GitHub repository</span>
            <span className="text-xs">↗</span>
          </a>
          <a
            href="https://github.com/elouan/dockyard/blob/main/README.md"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-between py-2 border-b last:border-0 hover:text-foreground text-muted-foreground transition-colors"
          >
            <span>Documentation</span>
            <span className="text-xs">↗</span>
          </a>
        </div>
      </SectionCard>
    </>
  );
}
