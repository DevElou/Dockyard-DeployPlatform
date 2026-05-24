"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createRuntimeTarget,
  disableRuntimeTarget,
  enableRuntimeTarget,
  listRuntimeTargets,
} from "@/lib/api/runtime-targets";
import { queryKeys } from "@/lib/query-keys";
import type { CreateRuntimeTargetPayload } from "@/lib/types/runtime-target";

export function useRuntimeTargets() {
  return useQuery({
    queryKey: queryKeys.runtimeTargets.list(),
    queryFn: listRuntimeTargets,
  });
}

export function useCreateRuntimeTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateRuntimeTargetPayload) =>
      createRuntimeTarget(payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.runtimeTargets.list() }),
  });
}

export function useEnableRuntimeTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => enableRuntimeTarget(id),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.runtimeTargets.list() }),
  });
}

export function useDisableRuntimeTarget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => disableRuntimeTarget(id),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.runtimeTargets.list() }),
  });
}
