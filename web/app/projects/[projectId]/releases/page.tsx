"use client";

import { use, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { BuildStatusBadge } from "@/components/status/build-status-badge";
import { useReleases, useCreateRelease } from "@/lib/hooks/use-releases";
import { useProject } from "@/lib/hooks/use-projects";
import { createReleaseSchema, type CreateReleaseValues } from "@/lib/validation/release";
import { formatRelative } from "@/lib/format";
import { handleMutationError } from "@/lib/api-error";

export default function ReleasesPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [open, setOpen] = useState(false);
  const { data: project } = useProject(projectId);
  const { data, isLoading, isError, refetch } = useReleases(projectId);
  const { mutate, isPending } = useCreateRelease(projectId);

  const form = useForm<CreateReleaseValues>({
    resolver: zodResolver(createReleaseSchema),
    defaultValues: { gitRef: project?.defaultBranch ?? "main" },
  });

  function onSubmit(values: CreateReleaseValues) {
    const version = `v${new Date().toISOString().slice(0, 10).replace(/-/g, ".")}.${Date.now().toString(36)}`;
    mutate(
      { gitRef: values.gitRef, version },
      {
        onSuccess: () => {
          toast.success("Release created — build started");
          setOpen(false);
          form.reset({ gitRef: project?.defaultBranch ?? "main" });
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  return (
    <div>
      <PageHeader title="Releases">
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="mr-1.5 h-4 w-4" />
          Refresh
        </Button>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          New release
        </Button>
      </PageHeader>

      {isLoading && <LoadingState />}
      {isError && (
        <ErrorState message="Could not load releases." onRetry={() => refetch()} />
      )}

      {!isLoading && !isError && data?.items.length === 0 && (
        <EmptyState
          title="No releases yet"
          description="Trigger a build from a Git ref to create your first release."
        >
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-1.5 h-4 w-4" />
            New release
          </Button>
        </EmptyState>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Version</TableHead>
              <TableHead>Ref</TableHead>
              <TableHead>Commit</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.items.map((r) => (
              <TableRow key={r.id}>
                <TableCell className="font-mono text-sm">{r.version}</TableCell>
                <TableCell className="font-mono text-sm">{r.gitRef}</TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {r.gitSha ? r.gitSha.slice(0, 7) : "—"}
                </TableCell>
                <TableCell>
                  <BuildStatusBadge status={r.buildStatus} />
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {formatRelative(r.createdAt)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New release</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="gitRef"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Git ref (branch, tag, or SHA)</FormLabel>
                    <FormControl>
                      <Input placeholder="main" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" onClick={() => setOpen(false)} type="button">
                  Cancel
                </Button>
                <Button type="submit" disabled={isPending}>
                  {isPending ? "Creating…" : "Create release"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
