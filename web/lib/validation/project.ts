import { z } from "zod";

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

export function parseGitHubUrl(url: string): { owner: string; repo: string } | null {
  try {
    const u = new URL(url);
    const parts = u.pathname.replace(/^\//, "").replace(/\.git$/, "").split("/");
    if (parts.length >= 2) {
      return { owner: parts[0], repo: parts[1] };
    }
  } catch {}
  return null;
}

export function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
