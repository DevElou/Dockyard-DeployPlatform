export type ProjectStatus = "active" | "archived";

export interface Project {
  id: string;
  slug: string;
  name: string;
  status: ProjectStatus;
  githubOwner: string;
  githubRepo: string;
  defaultBranch: string;
  rootDirectory: string;
  dockerfilePath: string;
  buildContext: string;
}

export interface CreateProjectPayload {
  slug: string;
  name: string;
  githubOwner: string;
  githubRepo: string;
  defaultBranch: string;
  rootDirectory: string;
  dockerfilePath: string;
  buildContext: string;
}
