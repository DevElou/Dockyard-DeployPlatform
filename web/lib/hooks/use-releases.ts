"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createRelease, getRelease, listReleases } from "@/lib/api/releases";
import { queryKeys } from "@/lib/query-keys";
import type { BuildStatus } from "@/lib/types/release";
import type { CreateReleasePayload } from "@/lib/types/release";

const IN_FLIGHT: BuildStatus[] = ["pending", "running"];

export function useReleases(projectId: string) {
  return useQuery({
    queryKey: queryKeys.releases.list(projectId),
    queryFn: () => listReleases(projectId),
    enabled: Boolean(projectId),
    refetchInterval: (query) => {
      const items = query.state.data?.items ?? [];
      return items.some((r) => IN_FLIGHT.includes(r.buildStatus)) ? 3000 : false;
    },
    refetchIntervalInBackground: false,
  });
}

export function useRelease(projectId: string, releaseId: string) {
  return useQuery({
    queryKey: queryKeys.releases.detail(projectId, releaseId),
    queryFn: () => getRelease(projectId, releaseId),
    enabled: Boolean(projectId) && Boolean(releaseId),
    refetchInterval: (query) => {
      const status = query.state.data?.buildStatus;
      return status && IN_FLIGHT.includes(status) ? 3000 : false;
    },
    refetchIntervalInBackground: false,
  });
}

export function useCreateRelease(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateReleasePayload) =>
      createRelease(projectId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.releases.list(projectId) }),
  });
}
