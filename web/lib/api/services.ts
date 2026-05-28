import { apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type { CreateServicePayload, ProjectService } from "@/lib/types/service";

export function listServices(
  projectId: string,
): Promise<ListResponse<ProjectService>> {
  return apiGet(`/api/v1/projects/${projectId}/services`);
}

export function getService(
  projectId: string,
  serviceId: string,
): Promise<ProjectService> {
  return apiGet(`/api/v1/projects/${projectId}/services/${serviceId}`);
}

export function createService(
  projectId: string,
  payload: CreateServicePayload,
): Promise<ProjectService> {
  return apiPost(`/api/v1/projects/${projectId}/services`, payload);
}
