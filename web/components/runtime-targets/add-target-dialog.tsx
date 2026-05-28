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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useCreateRuntimeTarget } from "@/lib/hooks/use-runtime-targets";
import {
  createRuntimeTargetSchema,
  type CreateRuntimeTargetValues,
} from "@/lib/validation/runtime-target";
import { handleMutationError } from "@/lib/api-error";
import { slugify } from "@/lib/utils";

interface AddTargetDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddTargetDialog({ open, onOpenChange }: AddTargetDialogProps) {
  const { mutate: create, isPending } = useCreateRuntimeTarget();

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
          onOpenChange(false);
          form.reset();
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
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
                  <FormDescription>URL of the deploy-agent on this host</FormDescription>
                  <FormControl>
                    <Input placeholder="http://deploy-agent:8090" {...field} />
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
              <Button variant="outline" type="button" onClick={() => onOpenChange(false)}>
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
  );
}
