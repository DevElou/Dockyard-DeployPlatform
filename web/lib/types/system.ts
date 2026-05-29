export interface IntegrationInfo {
  enabled: boolean;
  baseUrl?: string;
}

export interface SystemIntegrations {
  github: IntegrationInfo;
  npm: IntegrationInfo;
  dns: IntegrationInfo;
  registry: IntegrationInfo;
}

export interface SystemInfo {
  version: string;
  integrations: SystemIntegrations;
}
