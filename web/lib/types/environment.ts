export interface EnvironmentSet {
  id: string;
  projectId: string;
  name: string;
}

export interface EnvironmentVariable {
  id: string;
  environmentSetId: string;
  key: string;
  value: string;
  isSecret: boolean;
}

export interface CreateEnvironmentSetPayload {
  name: string;
}

export interface UpsertVariablePayload {
  key: string;
  value: string;
  isSecret: boolean;
}
