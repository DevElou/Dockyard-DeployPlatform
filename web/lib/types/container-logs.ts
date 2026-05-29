export interface ContainerLogs {
  deploymentId: string;
  containerId: string;
  tail: number;
  logs: string;
}
