import { apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type { ContainerLogs } from "@/lib/types/container-logs";
import type { OperationEvent } from "@/lib/types/operation-event";
import type {
  CreateDeploymentPayload,
  Deployment,
} from "@/lib/types/deployment";

export function listDeployments(
  projectId: string,
): Promise<ListResponse<Deployment>> {
  return apiGet(`/api/v1/projects/${projectId}/deployments`);
}

export function getDeployment(
  projectId: string,
  deploymentId: string,
): Promise<Deployment> {
  return apiGet(`/api/v1/projects/${projectId}/deployments/${deploymentId}`);
}

export function createDeployment(
  projectId: string,
  payload: CreateDeploymentPayload,
): Promise<Deployment> {
  return apiPost(`/api/v1/projects/${projectId}/deployments`, payload);
}

export function listDeploymentEvents(
  projectId: string,
  deploymentId: string,
): Promise<ListResponse<OperationEvent>> {
  return apiGet(
    `/api/v1/projects/${projectId}/deployments/${deploymentId}/events`,
  );
}

export function getDeploymentContainerLogs(
  projectId: string,
  deploymentId: string,
  tail?: number,
): Promise<ContainerLogs> {
  const params = tail && tail > 0 ? `?tail=${tail}` : "";
  return apiGet(
    `/api/v1/projects/${projectId}/deployments/${deploymentId}/container-logs${params}`,
  );
}
