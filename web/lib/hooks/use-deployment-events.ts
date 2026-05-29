import { useQuery } from "@tanstack/react-query";
import { listDeploymentEvents } from "@/lib/api/deployments";
import { queryKeys } from "@/lib/query-keys";
import type { DeploymentStatus } from "@/lib/types/deployment";

const IN_FLIGHT: DeploymentStatus[] = ["pending", "deploying"];

interface Options {
  enabled?: boolean;
  status?: DeploymentStatus;
}

export function useDeploymentEvents(
  projectId: string,
  deploymentId: string,
  { enabled = true, status }: Options = {},
) {
  return useQuery({
    queryKey: queryKeys.deployments.events(projectId, deploymentId),
    queryFn: () => listDeploymentEvents(projectId, deploymentId),
    enabled: enabled && Boolean(projectId) && Boolean(deploymentId),
    refetchInterval: status && IN_FLIGHT.includes(status) ? 2000 : false,
    refetchIntervalInBackground: false,
  });
}
