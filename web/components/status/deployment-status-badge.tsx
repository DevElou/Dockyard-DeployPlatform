import { Badge } from "@/components/ui/badge";
import type { DeploymentStatus } from "@/lib/types/deployment";

const config: Record<DeploymentStatus, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  pending: { label: "Pending", variant: "secondary" },
  deploying: { label: "Deploying", variant: "default" },
  healthy: { label: "Healthy", variant: "outline" },
  failed: { label: "Failed", variant: "destructive" },
  rolled_back: { label: "Rolled back", variant: "secondary" },
};

export function DeploymentStatusBadge({ status }: { status: DeploymentStatus }) {
  const { label, variant } = config[status] ?? { label: status, variant: "secondary" };
  return <Badge variant={variant}>{label}</Badge>;
}
