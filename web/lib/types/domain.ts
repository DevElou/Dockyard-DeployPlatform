export type DomainStatus = "pending" | "provisioning" | "ready" | "failed";

export interface Domain {
  id: string;
  projectId: string;
  projectServiceId: string | null;
  hostname: string;
  baseDomain: string;
  provider: string;
  routingType: string;
  tlsEnabled: boolean;
  status: DomainStatus;
}

export interface CreateDomainPayload {
  hostname: string;
  baseDomain: string;
  provider: string;
  routingType: string;
  tlsEnabled: boolean;
  projectServiceId?: string;
}
