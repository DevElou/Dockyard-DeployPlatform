"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
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
import { useCreateService } from "@/lib/hooks/use-services";
import { createServiceSchema, type CreateServiceValues } from "@/lib/validation/service";
import { handleMutationError } from "@/lib/api-error";

interface AddServiceDialogProps {
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddServiceDialog({ projectId, open, onOpenChange }: AddServiceDialogProps) {
  const { mutate, isPending } = useCreateService(projectId);

  const form = useForm<CreateServiceValues>({
    resolver: zodResolver(createServiceSchema),
    defaultValues: {
      name: "",
      containerPort: 3000,
      healthcheckPath: "/",
      healthcheckPort: 3000,
      routingEnabled: true,
    },
  });

  function onSubmit(values: CreateServiceValues) {
    mutate(values, {
      onSuccess: () => {
        toast.success("Service added");
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
          <DialogTitle>Add service</DialogTitle>
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
                    <Input placeholder="web" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-3">
              <FormField
                control={form.control}
                name="containerPort"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Container port</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        placeholder="3000"
                        {...field}
                        onChange={(e) => field.onChange(e.target.valueAsNumber)}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="healthcheckPath"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Healthcheck path</FormLabel>
                    <FormControl>
                      <Input placeholder="/" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="routingEnabled"
              render={({ field }) => (
                <FormItem className="flex items-center gap-2">
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                  <FormLabel className="!mt-0">Enable routing via NPM</FormLabel>
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Adding…" : "Add service"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
