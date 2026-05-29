export type OperationLevel = "info" | "warn" | "error" | "success";

export type OperationResourceType = "release" | "deployment";

export interface OperationEvent {
  id: string;
  resourceType: OperationResourceType;
  resourceId: string;
  phase: string;
  level: OperationLevel;
  message: string;
  details?: Record<string, string> | null;
  createdAt: string;
}
