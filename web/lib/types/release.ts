export type BuildStatus = "pending" | "running" | "succeeded" | "failed";

export interface Release {
  id: string;
  projectId: string;
  version: string;
  sourceType: string;
  gitSha: string | null;
  gitRef: string;
  imageRepository: string | null;
  imageTag: string | null;
  imageDigest: string | null;
  buildStatus: BuildStatus;
  createdAt: string;
}

export interface CreateReleasePayload {
  version: string;
  gitRef: string;
}
