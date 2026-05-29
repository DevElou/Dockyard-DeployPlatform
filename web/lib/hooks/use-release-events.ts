import { useQuery } from "@tanstack/react-query";
import { listReleaseEvents } from "@/lib/api/releases";
import { queryKeys } from "@/lib/query-keys";
import type { BuildStatus } from "@/lib/types/release";

const IN_FLIGHT: BuildStatus[] = ["pending", "running"];

interface Options {
  enabled?: boolean;
  buildStatus?: BuildStatus;
}

export function useReleaseEvents(
  projectId: string,
  releaseId: string,
  { enabled = true, buildStatus }: Options = {},
) {
  return useQuery({
    queryKey: queryKeys.releases.events(projectId, releaseId),
    queryFn: () => listReleaseEvents(projectId, releaseId),
    enabled: enabled && Boolean(projectId) && Boolean(releaseId),
    refetchInterval:
      buildStatus && IN_FLIGHT.includes(buildStatus) ? 2000 : false,
    refetchIntervalInBackground: false,
  });
}
