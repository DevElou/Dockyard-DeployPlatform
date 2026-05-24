import { apiDelete, apiGet, apiPost, apiPut } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type {
  CreateEnvironmentSetPayload,
  EnvironmentSet,
  EnvironmentVariable,
  UpsertVariablePayload,
} from "@/lib/types/environment";

export function listEnvironmentSets(
  projectId: string,
): Promise<ListResponse<EnvironmentSet>> {
  return apiGet(`/api/v1/projects/${projectId}/environments`);
}

export function createEnvironmentSet(
  projectId: string,
  payload: CreateEnvironmentSetPayload,
): Promise<EnvironmentSet> {
  return apiPost(`/api/v1/projects/${projectId}/environments`, payload);
}

export function listVariables(
  projectId: string,
  envId: string,
): Promise<ListResponse<EnvironmentVariable>> {
  return apiGet(
    `/api/v1/projects/${projectId}/environments/${envId}/variables`,
  );
}

export function upsertVariable(
  projectId: string,
  envId: string,
  payload: UpsertVariablePayload,
): Promise<EnvironmentVariable> {
  return apiPut(
    `/api/v1/projects/${projectId}/environments/${envId}/variables`,
    payload,
  );
}

export function deleteVariable(
  projectId: string,
  envId: string,
  varId: string,
): Promise<void> {
  return apiDelete(
    `/api/v1/projects/${projectId}/environments/${envId}/variables/${varId}`,
  );
}
