"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createEnvironmentSet,
  deleteVariable,
  listEnvironmentSets,
  listVariables,
  upsertVariable,
} from "@/lib/api/environments";
import { queryKeys } from "@/lib/query-keys";
import type {
  CreateEnvironmentSetPayload,
  UpsertVariablePayload,
} from "@/lib/types/environment";

export function useEnvironmentSets(projectId: string) {
  return useQuery({
    queryKey: queryKeys.environments.list(projectId),
    queryFn: () => listEnvironmentSets(projectId),
    enabled: Boolean(projectId),
  });
}

export function useCreateEnvironmentSet(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateEnvironmentSetPayload) =>
      createEnvironmentSet(projectId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.environments.list(projectId) }),
  });
}

export function useVariables(projectId: string, envId: string) {
  return useQuery({
    queryKey: queryKeys.environments.variables(projectId, envId),
    queryFn: () => listVariables(projectId, envId),
    enabled: Boolean(projectId) && Boolean(envId),
  });
}

export function useUpsertVariable(projectId: string, envId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: UpsertVariablePayload) =>
      upsertVariable(projectId, envId, payload),
    onSuccess: () =>
      qc.invalidateQueries({
        queryKey: queryKeys.environments.variables(projectId, envId),
      }),
  });
}

export function useDeleteVariable(projectId: string, envId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (varId: string) => deleteVariable(projectId, envId, varId),
    onSuccess: () =>
      qc.invalidateQueries({
        queryKey: queryKeys.environments.variables(projectId, envId),
      }),
  });
}
