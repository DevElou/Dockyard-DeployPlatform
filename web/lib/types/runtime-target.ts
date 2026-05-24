export type RuntimeType = "docker";

export interface RuntimeTarget {
  id: string;
  slug: string;
  name: string;
  runtimeType: RuntimeType;
  endpoint: string;
  agentKeyHash: string;
  serverGroup: string | null;
  region: string | null;
  enabled: boolean;
}

export interface CreateRuntimeTargetPayload {
  slug: string;
  name: string;
  runtimeType: RuntimeType;
  endpoint: string;
  agentKey: string;
  serverGroup?: string;
  region?: string;
}
