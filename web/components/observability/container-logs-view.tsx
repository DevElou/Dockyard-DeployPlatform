"use client";

import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ContainerLogs } from "@/lib/types/container-logs";

interface ContainerLogsViewProps {
  data?: ContainerLogs;
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
  onRefresh: () => void;
}

export function ContainerLogsView({
  data,
  isLoading,
  isError,
  errorMessage,
  onRefresh,
}: ContainerLogsViewProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-2">
        <div className="text-xs text-muted-foreground">
          {data ? (
            <>
              Container <code className="font-mono">{data.containerId.slice(0, 12)}</code>{" "}
              · last {data.tail} lines
            </>
          ) : (
            "Live snapshot from the deploy agent."
          )}
        </div>
        <Button variant="outline" size="sm" onClick={onRefresh} disabled={isLoading}>
          <RefreshCw className={`mr-1.5 h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {isError && (
        <p className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {errorMessage ?? "Could not load container logs."}
        </p>
      )}

      <pre
        className="max-h-[60vh] min-h-[12rem] overflow-auto rounded-md border bg-muted/40 p-3 text-xs font-mono whitespace-pre-wrap break-words"
        aria-busy={isLoading}
      >
        {data?.logs?.trim()
          ? data.logs
          : isLoading
            ? "Loading logs…"
            : "No log output."}
      </pre>
    </div>
  );
}
