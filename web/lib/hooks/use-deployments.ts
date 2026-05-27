import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createDeployment,
  listDeployments,
} from "@/lib/api/deployments";
import { queryKeys } from "@/lib/query-keys";
import type { CreateDeploymentPayload, DeploymentStatus } from "@/lib/types/deployment";

const IN_FLIGHT: DeploymentStatus[] = ["pending", "deploying"];

export function useDeployments(projectId: string) {
  return useQuery({
    queryKey: queryKeys.deployments.list(projectId),
    queryFn: () => listDeployments(projectId),
    enabled: Boolean(projectId),
    refetchInterval: (query) => {
      const items = query.state.data?.items ?? [];
      return items.some((d) => IN_FLIGHT.includes(d.status)) ? 3000 : false;
    },
    refetchIntervalInBackground: false,
  });
}

export function useCreateDeployment(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateDeploymentPayload) =>
      createDeployment(projectId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.deployments.list(projectId) }),
  });
}
