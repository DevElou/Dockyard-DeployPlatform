"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
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
import { useCreateDeployment } from "@/lib/hooks/use-deployments";
import { useReleases } from "@/lib/hooks/use-releases";
import { useRuntimeTargets } from "@/lib/hooks/use-runtime-targets";
import { createDeploymentSchema, type CreateDeploymentValues } from "@/lib/validation/deployment";
import { handleMutationError } from "@/lib/api-error";

interface NewDeploymentDialogProps {
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NewDeploymentDialog({ projectId, open, onOpenChange }: NewDeploymentDialogProps) {
  const { data: releasesData } = useReleases(projectId);
  const { data: targetsData } = useRuntimeTargets();
  const { mutate, isPending } = useCreateDeployment(projectId);

  const succeededReleases =
    releasesData?.items.filter((r) => r.buildStatus === "succeeded") ?? [];
  const enabledTargets = targetsData?.items.filter((t) => t.enabled) ?? [];

  const form = useForm<CreateDeploymentValues>({
    resolver: zodResolver(createDeploymentSchema),
    defaultValues: { releaseId: "", runtimeTargetId: "", strategy: "recreate" },
  });

  function onSubmit(values: CreateDeploymentValues) {
    mutate(values, {
      onSuccess: () => {
        toast.success("Deployment started");
        onOpenChange(false);
        form.reset();
      },
      onError: (err) => handleMutationError(err, form),
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
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
              <Button variant="outline" type="button" onClick={() => onOpenChange(false)}>
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
  );
}
