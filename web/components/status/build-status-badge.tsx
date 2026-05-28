import { Badge } from "@/components/ui/badge";
import type { BuildStatus } from "@/lib/types/release";

const config: Record<BuildStatus, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  pending: { label: "Pending", variant: "secondary" },
  running: { label: "Building", variant: "default" },
  succeeded: { label: "Succeeded", variant: "outline" },
  failed: { label: "Failed", variant: "destructive" },
};

export function BuildStatusBadge({ status }: { status: BuildStatus }) {
  const { label, variant } = config[status] ?? { label: status, variant: "secondary" };
  return <Badge variant={variant}>{label}</Badge>;
}
