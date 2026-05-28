import { z } from "zod";

export const createRuntimeTargetSchema = z.object({
  name: z.string().min(1, "Name is required").max(63, "Max 63 characters"),
  endpoint: z
    .string()
    .url("Must be a valid URL (e.g. http://server-1:8080)"),
  agentKey: z.string().min(1, "Agent key is required"),
  serverGroup: z.string().optional(),
  region: z.string().optional(),
});

export type CreateRuntimeTargetValues = z.infer<typeof createRuntimeTargetSchema>;
