export const queryKeys = {
  projects: {
    list: () => ["projects"] as const,
    detail: (id: string) => ["projects", id] as const,
    runtimeTargets: (id: string) => ["projects", id, "runtime-targets"] as const,
  },
  runtimeTargets: {
    list: () => ["runtime-targets"] as const,
    detail: (id: string) => ["runtime-targets", id] as const,
  },
  releases: {
    list: (projectId: string) => ["projects", projectId, "releases"] as const,
    detail: (projectId: string, releaseId: string) =>
      ["projects", projectId, "releases", releaseId] as const,
  },
  deployments: {
    list: (projectId: string) =>
      ["projects", projectId, "deployments"] as const,
    detail: (projectId: string, deploymentId: string) =>
      ["projects", projectId, "deployments", deploymentId] as const,
  },
  services: {
    list: (projectId: string) => ["projects", projectId, "services"] as const,
    detail: (projectId: string, serviceId: string) =>
      ["projects", projectId, "services", serviceId] as const,
  },
  environments: {
    list: (projectId: string) =>
      ["projects", projectId, "environments"] as const,
    variables: (projectId: string, envId: string) =>
      ["projects", projectId, "environments", envId, "variables"] as const,
  },
  domains: {
    list: (projectId: string) => ["projects", projectId, "domains"] as const,
  },
} as const;
