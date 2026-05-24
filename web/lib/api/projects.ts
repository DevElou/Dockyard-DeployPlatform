import { apiDelete, apiGet, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type { CreateProjectPayload, Project } from "@/lib/types/project";

export function listProjects(): Promise<ListResponse<Project>> {
  return apiGet("/api/v1/projects");
}

export function getProject(id: string): Promise<Project> {
  return apiGet(`/api/v1/projects/${id}`);
}

export function createProject(payload: CreateProjectPayload): Promise<Project> {
  return apiPost("/api/v1/projects", payload);
}

export function deleteProject(id: string): Promise<void> {
  return apiDelete(`/api/v1/projects/${id}`);
}

export function listProjectRuntimeTargets(
  projectId: string,
): Promise<ListResponse<{ id: string }>> {
  return apiGet(`/api/v1/projects/${projectId}/runtime-targets`);
}

export function addProjectRuntimeTarget(
  projectId: string,
  runtimeTargetId: string,
): Promise<void> {
  return apiPost(`/api/v1/projects/${projectId}/runtime-targets`, {
    runtimeTargetId,
  });
}
