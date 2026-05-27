import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createProject,
  deleteProject,
  getProject,
  listProjects,
} from "@/lib/api/projects";
import { queryKeys } from "@/lib/query-keys";
import type { CreateProjectPayload } from "@/lib/types/project";

export function useProjects() {
  return useQuery({
    queryKey: queryKeys.projects.list(),
    queryFn: listProjects,
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: queryKeys.projects.detail(id),
    queryFn: () => getProject(id),
    enabled: Boolean(id),
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateProjectPayload) => createProject(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.projects.list() }),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.projects.list() }),
  });
}
