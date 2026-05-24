export type DeploymentStatus =
  | "pending"
  | "deploying"
  | "healthy"
  | "failed"
  | "rolled_back";

export interface Deployment {
  id: string;
  projectId: string;
  releaseId: string;
  runtimeTargetId: string;
  projectServiceId: string | null;
  environmentSetId: string | null;
  status: DeploymentStatus;
  strategy: string;
  triggeredByUserId: string | null;
  rollbackOfDeploymentId: string | null;
  startedAt: string | null;
  finishedAt: string | null;
  createdAt: string;
}

export interface CreateDeploymentPayload {
  releaseId: string;
  runtimeTargetId: string;
  strategy: string;
}
