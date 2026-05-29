"use client";

import { SectionCard } from "@/components/settings/section-card";
import { IntegrationStatusRow } from "@/components/settings/integration-status-row";
import { useSystemInfo } from "@/lib/hooks/use-system";
import { Skeleton } from "@/components/ui/skeleton";

export default function IntegrationsPage() {
  const { data, isPending } = useSystemInfo();

  if (isPending) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-40 w-full rounded-lg" />
        <Skeleton className="h-40 w-full rounded-lg" />
      </div>
    );
  }

  const { github, npm, dns, registry } = data?.integrations ?? {
    github: { enabled: false },
    npm: { enabled: false },
    dns: { enabled: false },
    registry: { enabled: false },
  };

  return (
    <>
      <SectionCard
        title="Source & CI"
        description="Integrations for fetching source code and building images."
      >
        <IntegrationStatusRow
          name="GitHub"
          description="Source code provider. Configured via DOCKYARD_GITHUB_TOKEN environment variable."
          enabled={github.enabled}
        />
        <IntegrationStatusRow
          name="Container Registry"
          description="Stores built Docker images. Configured via DOCKYARD_REGISTRY_URL."
          enabled={registry.enabled}
          meta={registry.baseUrl ? [{ label: "URL", value: registry.baseUrl }] : undefined}
        />
      </SectionCard>

      <SectionCard
        title="Routing & DNS"
        description="Integrations for exposing services and managing domain names."
      >
        <IntegrationStatusRow
          name="Nginx Proxy Manager"
          description="Reverse proxy for routing traffic to deployed services. Configured via DOCKYARD_NPM_URL."
          enabled={npm.enabled}
          meta={npm.baseUrl ? [{ label: "Base URL", value: npm.baseUrl }] : undefined}
        />
        <IntegrationStatusRow
          name="DuckDNS"
          description="Dynamic DNS provider. Not yet wired — coming in a future release."
          enabled={dns.enabled}
        />
      </SectionCard>
    </>
  );
}
