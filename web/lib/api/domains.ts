import { apiDelete, apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type { CreateDomainPayload, Domain } from "@/lib/types/domain";

export function listDomains(
  projectId: string,
): Promise<ListResponse<Domain>> {
  return apiGet(`/api/v1/projects/${projectId}/domains`);
}

export function getDomain(
  projectId: string,
  domainId: string,
): Promise<Domain> {
  return apiGet(`/api/v1/projects/${projectId}/domains/${domainId}`);
}

export function createDomain(
  projectId: string,
  payload: CreateDomainPayload,
): Promise<Domain> {
  return apiPost(`/api/v1/projects/${projectId}/domains`, payload);
}

export function deleteDomain(
  projectId: string,
  domainId: string,
): Promise<void> {
  return apiDelete(`/api/v1/projects/${projectId}/domains/${domainId}`);
}
