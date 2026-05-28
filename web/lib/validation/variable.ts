import { z } from "zod";

export const upsertVariableSchema = z.object({
  key: z
    .string()
    .min(1, "Key is required")
    .regex(/^[A-Z0-9_]+$/, "Key must be uppercase letters, digits, or underscores"),
  value: z.string(),
  isSecret: z.boolean(),
});

export const createEnvironmentSetSchema = z.object({
  name: z.string().min(1, "Name is required").max(63, "Max 63 characters"),
});

export type UpsertVariableValues = z.infer<typeof upsertVariableSchema>;
export type CreateEnvironmentSetValues = z.infer<typeof createEnvironmentSetSchema>;
