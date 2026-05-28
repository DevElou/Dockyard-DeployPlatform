"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, Plus, Trash2 } from "lucide-react";
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
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { useVariables, useUpsertVariable, useDeleteVariable } from "@/lib/hooks/use-environments";
import { upsertVariableSchema, type UpsertVariableValues } from "@/lib/validation/variable";
import { handleMutationError } from "@/lib/api-error";
import type { EnvironmentSet } from "@/lib/types/environment";

interface VariableListProps {
  projectId: string;
  envSet: EnvironmentSet;
}

export function VariableList({ projectId, envSet }: VariableListProps) {
  const [addOpen, setAddOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [revealed, setRevealed] = useState<Set<string>>(new Set());

  const { data } = useVariables(projectId, envSet.id);
  const { mutate: upsert, isPending: upserting } = useUpsertVariable(projectId, envSet.id);
  const { mutate: del, isPending: deleting } = useDeleteVariable(projectId, envSet.id);

  const varForm = useForm<UpsertVariableValues>({
    resolver: zodResolver(upsertVariableSchema),
    defaultValues: { key: "", value: "", isSecret: false },
  });

  function toggleReveal(id: string) {
    setRevealed((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  function onUpsert(values: UpsertVariableValues) {
    upsert(values, {
      onSuccess: () => {
        toast.success("Variable saved");
        setAddOpen(false);
        varForm.reset();
      },
      onError: (err) => handleMutationError(err, varForm),
    });
  }

  const vars = data?.items ?? [];

  return (
    <div>
      <div className="space-y-1 mb-2">
        {vars.map((v) => (
          <div
            key={v.id}
            className="flex items-center justify-between rounded bg-muted/50 px-3 py-2"
          >
            <div className="min-w-0 flex-1">
              <span className="font-mono text-xs font-medium">{v.key}</span>
              <span className="mx-2 text-muted-foreground">=</span>
              {v.isSecret && !revealed.has(v.id) ? (
                <span className="font-mono text-xs text-muted-foreground">••••••••</span>
              ) : (
                <span className="font-mono text-xs">{v.value}</span>
              )}
            </div>
            <div className="flex items-center gap-1 ml-2">
              {v.isSecret && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={() => toggleReveal(v.id)}
                >
                  {revealed.has(v.id) ? (
                    <EyeOff className="h-3 w-3" />
                  ) : (
                    <Eye className="h-3 w-3" />
                  )}
                </Button>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 text-destructive"
                onClick={() => setDeleteId(v.id)}
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>
          </div>
        ))}
      </div>

      <Button variant="ghost" size="sm" onClick={() => setAddOpen(true)}>
        <Plus className="mr-1.5 h-3.5 w-3.5" />
        Add variable
      </Button>

      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add variable — {envSet.name}</DialogTitle>
          </DialogHeader>
          <Form {...varForm}>
            <form onSubmit={varForm.handleSubmit(onUpsert)} className="space-y-4">
              <FormField
                control={varForm.control}
                name="key"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Key</FormLabel>
                    <FormControl>
                      <Input placeholder="DATABASE_URL" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={varForm.control}
                name="value"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Value</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={varForm.control}
                name="isSecret"
                render={({ field }) => (
                  <FormItem className="flex items-center gap-2">
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                    <FormLabel className="!mt-0">Secret (masked in UI)</FormLabel>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => setAddOpen(false)}>
                  Cancel
                </Button>
                <Button type="submit" disabled={upserting}>
                  {upserting ? "Saving…" : "Save"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Delete variable?"
        description="This action cannot be undone."
        confirmLabel="Delete"
        destructive
        loading={deleting}
        onConfirm={() => {
          if (deleteId) {
            del(deleteId, {
              onSuccess: () => {
                toast.success("Variable deleted");
                setDeleteId(null);
              },
            });
          }
        }}
      />
    </div>
  );
}
