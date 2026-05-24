"use client";

import { use, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { DeploymentStatusBadge } from "@/components/status/deployment-status-badge";
import { useDeployments, useCreateDeployment } from "@/lib/hooks/use-deployments";
import { useReleases } from "@/lib/hooks/use-releases";
import { useRuntimeTargets } from "@/lib/hooks/use-runtime-targets";
import { createDeploymentSchema, type CreateDeploymentValues } from "@/lib/validation/deployment";
import { formatRelative } from "@/lib/format";
import { handleMutationError } from "@/lib/api-error";
import { Rocket } from "lucide-react";

export default function DeploymentsPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [open, setOpen] = useState(false);
  const { data, isLoading, isError, refetch } = useDeployments(projectId);
  const { data: releasesData } = useReleases(projectId);
  const { data: targetsData } = useRuntimeTargets();
  const { mutate, isPending } = useCreateDeployment(projectId);

  const form = useForm<CreateDeploymentValues>({
    resolver: zodResolver(createDeploymentSchema),
    defaultValues: { releaseId: "", runtimeTargetId: "", strategy: "recreate" },
  });

  const succeededReleases = releasesData?.items.filter(
    (r) => r.buildStatus === "succeeded"
  ) ?? [];
  const enabledTargets = targetsData?.items.filter((t) => t.enabled) ?? [];

  function onSubmit(values: CreateDeploymentValues) {
    mutate(values, {
      onSuccess: () => {
        toast.success("Deployment started");
        setOpen(false);
        form.reset();
      },
      onError: (err) => handleMutationError(err, form),
    });
  }

  return (
    <div>
      <PageHeader title="Deployments">
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="mr-1.5 h-4 w-4" />
          Refresh
        </Button>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Rocket className="mr-1.5 h-4 w-4" />
          Deploy
        </Button>
      </PageHeader>

      {isLoading && <LoadingState />}
      {isError && (
        <ErrorState message="Could not load deployments." onRetry={() => refetch()} />
      )}

      {!isLoading && !isError && data?.items.length === 0 && (
        <EmptyState
          icon={Rocket}
          title="No deployments yet"
          description="Choose a release and runtime target to deploy."
        >
          <Button size="sm" onClick={() => setOpen(true)}>
            <Rocket className="mr-1.5 h-4 w-4" />
            Deploy
          </Button>
        </EmptyState>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Release</TableHead>
              <TableHead>Target</TableHead>
              <TableHead>Strategy</TableHead>
              <TableHead>Started</TableHead>
              <TableHead>Finished</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.items.map((d) => (
              <TableRow key={d.id}>
                <TableCell>
                  <DeploymentStatusBadge status={d.status} />
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {d.releaseId.slice(0, 8)}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {d.runtimeTargetId.slice(0, 8)}
                </TableCell>
                <TableCell className="text-sm">{d.strategy}</TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {d.startedAt ? formatRelative(d.startedAt) : "—"}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {d.finishedAt ? formatRelative(d.finishedAt) : "—"}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New deployment</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="releaseId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Release</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a release" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {succeededReleases.map((r) => (
                          <SelectItem key={r.id} value={r.id}>
                            {r.version} — {r.gitRef}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="runtimeTargetId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Runtime target</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a target" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {enabledTargets.map((t) => (
                          <SelectItem key={t.id} value={t.id}>
                            {t.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="strategy"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Strategy</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="recreate">Recreate</SelectItem>
                        <SelectItem value="rolling">Rolling</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setOpen(false)}
                  type="button"
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isPending}>
                  {isPending ? "Deploying…" : "Deploy"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
