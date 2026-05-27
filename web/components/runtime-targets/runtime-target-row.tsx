"use client";

import { CheckCircle2, XCircle } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  useEnableRuntimeTarget,
  useDisableRuntimeTarget,
} from "@/lib/hooks/use-runtime-targets";
import type { RuntimeTarget } from "@/lib/types/runtime-target";

interface RuntimeTargetRowProps {
  target: RuntimeTarget;
}

export function RuntimeTargetRow({ target }: RuntimeTargetRowProps) {
  const { mutate: enable } = useEnableRuntimeTarget();
  const { mutate: disable } = useDisableRuntimeTarget();

  function toggleEnabled() {
    if (target.enabled) {
      disable(target.id, {
        onSuccess: () => toast.success("Target disabled"),
        onError: () => toast.error("Failed to update target"),
      });
    } else {
      enable(target.id, {
        onSuccess: () => toast.success("Target enabled"),
        onError: () => toast.error("Failed to update target"),
      });
    }
  }

  return (
    <div className="flex items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-3">
        <div className={target.enabled ? "text-green-500" : "text-muted-foreground"}>
          {target.enabled ? (
            <CheckCircle2 className="h-5 w-5" />
          ) : (
            <XCircle className="h-5 w-5" />
          )}
        </div>
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{target.name}</span>
            <Badge variant="secondary" className="text-xs">
              {target.runtimeType}
            </Badge>
            {target.serverGroup && (
              <Badge variant="outline" className="text-xs">
                {target.serverGroup}
              </Badge>
            )}
          </div>
          <p className="mt-0.5 font-mono text-xs text-muted-foreground">
            {target.endpoint}
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <span className="text-xs text-muted-foreground">
          {target.enabled ? "Enabled" : "Disabled"}
        </span>
        <Switch checked={target.enabled} onCheckedChange={toggleEnabled} />
      </div>
    </div>
  );
}
