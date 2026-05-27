"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
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
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { useDomains, useCreateDomain, useDeleteDomain } from "@/lib/hooks/use-domains";
import { createDomainSchema, type CreateDomainValues } from "@/lib/validation/domain";
import { handleMutationError } from "@/lib/api-error";

interface DomainSectionProps {
  projectId: string;
  serviceId: string;
}

export function DomainSection({ projectId, serviceId }: DomainSectionProps) {
  const [addOpen, setAddOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const { data } = useDomains(projectId);
  const { mutate: create, isPending: creating } = useCreateDomain(projectId);
  const { mutate: del, isPending: deleting } = useDeleteDomain(projectId);

  const form = useForm<CreateDomainValues>({
    resolver: zodResolver(createDomainSchema),
    defaultValues: {
      hostname: "",
      baseDomain: "",
      provider: "manual",
      routingType: "proxy",
      tlsEnabled: false,
      projectServiceId: serviceId,
    },
  });

  function onSubmit(values: CreateDomainValues) {
    const baseDomain = values.hostname.split(".").slice(-2).join(".");
    create(
      { ...values, baseDomain, projectServiceId: serviceId },
      {
        onSuccess: () => {
          toast.success("Domain added");
          setAddOpen(false);
          form.reset();
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  const domains = (data?.items ?? []).filter((d) => d.projectServiceId === serviceId);

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <p className="text-sm font-medium">Domains</p>
        <Button variant="outline" size="sm" onClick={() => setAddOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          Add domain
        </Button>
      </div>

      {domains.length === 0 && (
        <p className="text-xs text-muted-foreground">No domains yet.</p>
      )}

      <div className="space-y-1">
        {domains.map((d) => (
          <div
            key={d.id}
            className="flex items-center justify-between rounded bg-muted/50 px-3 py-2"
          >
            <div className="min-w-0">
              <span className="font-mono text-xs font-medium">{d.hostname}</span>
              <Badge variant="secondary" className="ml-2 text-xs">
                {d.status}
              </Badge>
              {d.tlsEnabled && (
                <Badge variant="outline" className="ml-1 text-xs">
                  TLS
                </Badge>
              )}
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-destructive"
              onClick={() => setDeleteId(d.id)}
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        ))}
      </div>

      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add domain</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="hostname"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Hostname</FormLabel>
                    <FormControl>
                      <Input placeholder="app.example.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="tlsEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center gap-2">
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                    <FormLabel className="!mt-0">Enable TLS</FormLabel>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => setAddOpen(false)}>
                  Cancel
                </Button>
                <Button type="submit" disabled={creating}>
                  {creating ? "Adding…" : "Add domain"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Remove domain?"
        description="This will remove the domain and its routing configuration."
        confirmLabel="Remove"
        destructive
        loading={deleting}
        onConfirm={() => {
          if (deleteId) {
            del(deleteId, {
              onSuccess: () => {
                toast.success("Domain removed");
                setDeleteId(null);
              },
            });
          }
        }}
      />
    </div>
  );
}
