"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Server, CheckCircle2, XCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import {
  useRuntimeTargets,
  useCreateRuntimeTarget,
  useEnableRuntimeTarget,
  useDisableRuntimeTarget,
} from "@/lib/hooks/use-runtime-targets";
import {
  createRuntimeTargetSchema,
  type CreateRuntimeTargetValues,
} from "@/lib/validation/runtime-target";
import { handleMutationError } from "@/lib/api-error";
import { slugify } from "@/lib/validation/project";

export default function RuntimeTargetsPage() {
  const [open, setOpen] = useState(false);
  const { data, isLoading, isError, refetch } = useRuntimeTargets();
  const { mutate: create, isPending } = useCreateRuntimeTarget();
  const { mutate: enable } = useEnableRuntimeTarget();
  const { mutate: disable } = useDisableRuntimeTarget();

  const form = useForm<CreateRuntimeTargetValues>({
    resolver: zodResolver(createRuntimeTargetSchema),
    defaultValues: {
      name: "",
      endpoint: "",
      agentKey: "",
      serverGroup: "",
      region: "",
    },
  });

  function onSubmit(values: CreateRuntimeTargetValues) {
    create(
      {
        slug: slugify(values.name),
        name: values.name,
        runtimeType: "docker",
        endpoint: values.endpoint,
        agentKey: values.agentKey,
        serverGroup: values.serverGroup || undefined,
        region: values.region || undefined,
      },
      {
        onSuccess: () => {
          toast.success("Runtime target created");
          setOpen(false);
          form.reset();
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  function toggleEnabled(id: string, enabled: boolean) {
    if (enabled) {
      disable(id, {
        onSuccess: () => toast.success("Target disabled"),
        onError: () => toast.error("Failed to update target"),
      });
    } else {
      enable(id, {
        onSuccess: () => toast.success("Target enabled"),
        onError: () => toast.error("Failed to update target"),
      });
    }
  }

  return (
    <div>
      <PageHeader
        title="Runtime targets"
        description="Manage Docker hosts where deployments run"
      >
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          Add target
        </Button>
      </PageHeader>

      {isLoading && <LoadingState rows={3} />}
      {isError && (
        <ErrorState
          message="Could not load runtime targets."
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && data?.items.length === 0 && (
        <EmptyState
          icon={Server}
          title="No runtime targets"
          description="Add a Docker host with a running deploy-agent to get started."
        >
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-1.5 h-4 w-4" />
            Add target
          </Button>
        </EmptyState>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <div className="p-6 space-y-3">
          {data.items.map((target) => (
            <div
              key={target.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div className="flex items-center gap-3">
                <div
                  className={
                    target.enabled ? "text-green-500" : "text-muted-foreground"
                  }
                >
                  {target.enabled ? (
                    <CheckCircle2 className="h-5 w-5" />
                  ) : (
                    <XCircle className="h-5 w-5" />
                  )}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{target.name}</span>
                    <Badge variant="secondary" className="text-xs">
                      {target.runtimeType}
                    </Badge>
                    {target.serverGroup && (
                      <Badge variant="outline" className="text-xs">
                        {target.serverGroup}
                      </Badge>
                    )}
                  </div>
                  <p className="mt-0.5 font-mono text-xs text-muted-foreground">
                    {target.endpoint}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <span className="text-xs text-muted-foreground">
                  {target.enabled ? "Enabled" : "Disabled"}
                </span>
                <Switch
                  checked={target.enabled}
                  onCheckedChange={() => toggleEnabled(target.id, target.enabled)}
                />
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add runtime target</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="server-1" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="endpoint"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Agent endpoint</FormLabel>
                    <FormDescription>
                      URL of the deploy-agent on this host
                    </FormDescription>
                    <FormControl>
                      <Input placeholder="http://server-1:8080" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="agentKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Agent key</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="••••••••" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-2 gap-3">
                <FormField
                  control={form.control}
                  name="serverGroup"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Server group</FormLabel>
                      <FormControl>
                        <Input placeholder="homelab" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="region"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Region</FormLabel>
                      <FormControl>
                        <Input placeholder="eu-west" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setOpen(false)}
                  type="button"
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={isPending}>
                  {isPending ? "Adding…" : "Add target"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
