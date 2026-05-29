import { apiGet } from "./client";
import type { SystemInfo } from "@/lib/types/system";

export function getSystemInfo(): Promise<SystemInfo> {
  return apiGet("/api/v1/system/info");
}

export async function getHealth(): Promise<boolean> {
  try {
    await apiGet("/healthz");
    return true;
  } catch {
    return false;
  }
}
