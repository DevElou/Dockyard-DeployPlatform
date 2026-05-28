import { z } from "zod";

export const createDeploymentSchema = z.object({
  releaseId: z.string().min(1, "Release is required"),
  runtimeTargetId: z.string().min(1, "Runtime target is required"),
  strategy: z.enum(["recreate", "rolling"]),
});

export type CreateDeploymentValues = z.infer<typeof createDeploymentSchema>;
