import { apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
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
