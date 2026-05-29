import { useQuery } from "@tanstack/react-query";
import { getHealth, getSystemInfo } from "@/lib/api/system";
import { queryKeys } from "@/lib/query-keys";

export function useSystemInfo() {
  return useQuery({
    queryKey: queryKeys.system.info(),
    queryFn: getSystemInfo,
  });
}

export function useHealth() {
  return useQuery({
    queryKey: queryKeys.system.health(),
    queryFn: getHealth,
    refetchInterval: 30_000,
  });
}
