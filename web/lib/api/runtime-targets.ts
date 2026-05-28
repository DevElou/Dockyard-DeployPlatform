import { apiGet, apiPatch, apiPost } from "./client";
import type { ListResponse } from "@/lib/types/api";
import type {
  CreateRuntimeTargetPayload,
  RuntimeTarget,
} from "@/lib/types/runtime-target";

export function listRuntimeTargets(): Promise<ListResponse<RuntimeTarget>> {
  return apiGet("/api/v1/runtime-targets");
}

export function getRuntimeTarget(id: string): Promise<RuntimeTarget> {
  return apiGet(`/api/v1/runtime-targets/${id}`);
}

export function createRuntimeTarget(
  payload: CreateRuntimeTargetPayload,
): Promise<RuntimeTarget> {
  return apiPost("/api/v1/runtime-targets", payload);
}

export function enableRuntimeTarget(id: string): Promise<RuntimeTarget> {
  return apiPatch(`/api/v1/runtime-targets/${id}/enable`);
}

export function disableRuntimeTarget(id: string): Promise<RuntimeTarget> {
  return apiPatch(`/api/v1/runtime-targets/${id}/disable`);
}
