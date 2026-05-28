"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { EnvSection } from "./env-section";
import { DomainSection } from "./domain-section";
import type { ProjectService } from "@/lib/types/service";

interface ServiceCardProps {
  projectId: string;
  service: ProjectService;
}

export function ServiceCard({ projectId, service }: ServiceCardProps) {
  const [showDetails, setShowDetails] = useState(false);

  return (
    <div className="rounded-lg border">
      <div className="flex items-center justify-between p-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{service.name}</span>
            <Badge variant="secondary">:{service.containerPort}</Badge>
            {service.routingEnabled && (
              <Badge variant="outline">routing</Badge>
            )}
          </div>
          <p className="mt-0.5 text-xs text-muted-foreground">
            healthcheck: {service.healthcheckPath}:{service.healthcheckPort}
          </p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowDetails((v) => !v)}
        >
          {showDetails ? (
            <ChevronUp className="h-4 w-4" />
          ) : (
            <ChevronDown className="h-4 w-4" />
          )}
        </Button>
      </div>

      {showDetails && (
        <>
          <Separator />
          <div className="p-4 space-y-4">
            <EnvSection projectId={projectId} />
            <Separator />
            <DomainSection projectId={projectId} serviceId={service.id} />
          </div>
        </>
      )}
    </div>
  );
}
