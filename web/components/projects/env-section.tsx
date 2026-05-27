"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus } from "lucide-react";
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
import { VariableList } from "./variable-list";
import {
  useEnvironmentSets,
  useCreateEnvironmentSet,
} from "@/lib/hooks/use-environments";
import {
  createEnvironmentSetSchema,
  type CreateEnvironmentSetValues,
} from "@/lib/validation/variable";
import { handleMutationError } from "@/lib/api-error";

interface EnvSectionProps {
  projectId: string;
}

export function EnvSection({ projectId }: EnvSectionProps) {
  const [addSetOpen, setAddSetOpen] = useState(false);
  const [selectedSetId, setSelectedSetId] = useState<string | null>(null);

  const { data } = useEnvironmentSets(projectId);
  const { mutate: createSet, isPending: creatingSet } = useCreateEnvironmentSet(projectId);

  const setForm = useForm<CreateEnvironmentSetValues>({
    resolver: zodResolver(createEnvironmentSetSchema),
    defaultValues: { name: "" },
  });

  function onCreateSet(values: CreateEnvironmentSetValues) {
    createSet(values, {
      onSuccess: () => {
        toast.success("Environment set created");
        setAddSetOpen(false);
        setForm.reset();
      },
      onError: (err) => handleMutationError(err, setForm),
    });
  }

  const sets = data?.items ?? [];
  const activeSet = selectedSetId
    ? sets.find((s) => s.id === selectedSetId)
    : sets[0];

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <p className="text-sm font-medium">Environment variables</p>
        <Button variant="outline" size="sm" onClick={() => setAddSetOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          New set
        </Button>
      </div>

      {sets.length > 0 && (
        <div className="flex gap-1 mb-3 flex-wrap">
          {sets.map((s) => (
            <button
              key={s.id}
              onClick={() => setSelectedSetId(s.id)}
              className={`text-xs px-2 py-1 rounded border transition-colors ${
                activeSet?.id === s.id
                  ? "bg-primary text-primary-foreground border-primary"
                  : "border-border text-muted-foreground hover:text-foreground"
              }`}
            >
              {s.name}
            </button>
          ))}
        </div>
      )}

      {activeSet && <VariableList projectId={projectId} envSet={activeSet} />}

      {sets.length === 0 && (
        <p className="text-xs text-muted-foreground">No environment sets yet.</p>
      )}

      <Dialog open={addSetOpen} onOpenChange={setAddSetOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New environment set</DialogTitle>
          </DialogHeader>
          <Form {...setForm}>
            <form onSubmit={setForm.handleSubmit(onCreateSet)} className="space-y-4">
              <FormField
                control={setForm.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="production" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => setAddSetOpen(false)}>
                  Cancel
                </Button>
                <Button type="submit" disabled={creatingSet}>
                  {creatingSet ? "Creating…" : "Create"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
