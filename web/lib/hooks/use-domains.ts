"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createDomain, deleteDomain, listDomains } from "@/lib/api/domains";
import { queryKeys } from "@/lib/query-keys";
import type { CreateDomainPayload } from "@/lib/types/domain";

export function useDomains(projectId: string) {
  return useQuery({
    queryKey: queryKeys.domains.list(projectId),
    queryFn: () => listDomains(projectId),
    enabled: Boolean(projectId),
  });
}

export function useCreateDomain(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateDomainPayload) =>
      createDomain(projectId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.domains.list(projectId) }),
  });
}

export function useDeleteDomain(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (domainId: string) => deleteDomain(projectId, domainId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.domains.list(projectId) }),
  });
}
