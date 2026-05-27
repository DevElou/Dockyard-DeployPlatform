import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createService, listServices } from "@/lib/api/services";
import { queryKeys } from "@/lib/query-keys";
import type { CreateServicePayload } from "@/lib/types/service";

export function useServices(projectId: string) {
  return useQuery({
    queryKey: queryKeys.services.list(projectId),
    queryFn: () => listServices(projectId),
    enabled: Boolean(projectId),
  });
}

export function useCreateService(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateServicePayload) =>
      createService(projectId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.services.list(projectId) }),
  });
}
