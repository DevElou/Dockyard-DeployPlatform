"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { useCreateRelease } from "@/lib/hooks/use-releases";
import { useProject } from "@/lib/hooks/use-projects";
import { createReleaseSchema, type CreateReleaseValues } from "@/lib/validation/release";
import { handleMutationError } from "@/lib/api-error";

interface NewReleaseDialogProps {
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NewReleaseDialog({ projectId, open, onOpenChange }: NewReleaseDialogProps) {
  const { data: project } = useProject(projectId);
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
          onOpenChange(false);
          form.reset({ gitRef: project?.defaultBranch ?? "main" });
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
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
              <Button variant="outline" type="button" onClick={() => onOpenChange(false)}>
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
  );
}
