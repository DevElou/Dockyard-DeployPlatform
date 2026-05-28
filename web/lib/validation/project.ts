import { z } from "zod";

// slugify and parseGitHubUrl live in lib/utils — re-exported here for convenience.
export { slugify, parseGitHubUrl } from "@/lib/utils";

export const createProjectSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .max(63, "Max 63 characters"),
  repoUrl: z
    .string()
    .url("Must be a valid URL")
    .regex(/github\.com/, "Must be a GitHub repository URL"),
  defaultBranch: z.string().min(1, "Branch is required"),
  rootDirectory: z.string().min(1),
  dockerfilePath: z.string().min(1),
  buildContext: z.string().min(1),
});

export type CreateProjectValues = z.infer<typeof createProjectSchema>;
