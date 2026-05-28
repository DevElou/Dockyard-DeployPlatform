export interface ProjectService {
  id: string;
  projectId: string;
  name: string;
  containerPort: number;
  healthcheckPath: string;
  healthcheckPort: number;
  routingEnabled: boolean;
}

export interface CreateServicePayload {
  name: string;
  containerPort: number;
  healthcheckPath: string;
  healthcheckPort: number;
  routingEnabled: boolean;
}
