import { z } from "zod";

export const createReleaseSchema = z.object({
  gitRef: z.string().min(1, "Git ref is required"),
});

export type CreateReleaseValues = z.infer<typeof createReleaseSchema>;
