import { z } from "zod";

export const createServiceSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .max(63, "Max 63 characters")
    .regex(
      /^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$/,
      "Lowercase alphanumeric with hyphens"
    ),
  containerPort: z.number().int().min(1).max(65535),
  healthcheckPath: z.string().min(1),
  healthcheckPort: z.number().int().min(1).max(65535),
  routingEnabled: z.boolean(),
});

export type CreateServiceValues = z.infer<typeof createServiceSchema>;
