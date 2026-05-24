import { apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type { CreateReleasePayload, Release } from "@/lib/types/release";

export function listReleases(
  projectId: string,
): Promise<ListResponse<Release>> {
  return apiGet(`/api/v1/projects/${projectId}/releases`);
}

export function getRelease(
  projectId: string,
  releaseId: string,
): Promise<Release> {
  return apiGet(`/api/v1/projects/${projectId}/releases/${releaseId}`);
}

export function createRelease(
  projectId: string,
  payload: CreateReleasePayload,
): Promise<Release> {
  return apiPost(`/api/v1/projects/${projectId}/releases`, payload);
}
