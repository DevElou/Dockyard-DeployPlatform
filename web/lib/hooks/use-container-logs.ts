import { useQuery } from "@tanstack/react-query";
import { getDeploymentContainerLogs } from "@/lib/api/deployments";
import { queryKeys } from "@/lib/query-keys";

interface Options {
  enabled?: boolean;
  tail?: number;
}

export function useContainerLogs(
  projectId: string,
  deploymentId: string,
  { enabled = true, tail = 300 }: Options = {},
) {
  return useQuery({
    queryKey: [
      ...queryKeys.deployments.containerLogs(projectId, deploymentId),
      tail,
    ],
    queryFn: () => getDeploymentContainerLogs(projectId, deploymentId, tail),
    enabled: enabled && Boolean(projectId) && Boolean(deploymentId),
    // Logs are intentionally on-demand: the user opens the tab, sees the snapshot,
    // and can refetch manually. No background polling — they're large and noisy.
    refetchOnWindowFocus: false,
    staleTime: 0,
  });
}
