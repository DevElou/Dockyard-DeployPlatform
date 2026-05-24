"use client";

import { use, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Trash2, Eye, EyeOff, ChevronDown, ChevronUp } from "lucide-react";
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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/layout/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { useServices, useCreateService } from "@/lib/hooks/use-services";
import {
  useEnvironmentSets,
  useCreateEnvironmentSet,
  useVariables,
  useUpsertVariable,
  useDeleteVariable,
} from "@/lib/hooks/use-environments";
import { useDomains, useCreateDomain, useDeleteDomain } from "@/lib/hooks/use-domains";
import { createServiceSchema, type CreateServiceValues } from "@/lib/validation/service";
import {
  upsertVariableSchema,
  createEnvironmentSetSchema,
  type UpsertVariableValues,
  type CreateEnvironmentSetValues,
} from "@/lib/validation/variable";
import { createDomainSchema, type CreateDomainValues } from "@/lib/validation/domain";
import { handleMutationError } from "@/lib/api-error";
import type { ProjectService } from "@/lib/types/service";
import type { EnvironmentSet } from "@/lib/types/environment";

export default function ServicesPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = use(params);
  const [addServiceOpen, setAddServiceOpen] = useState(false);
  const { data, isLoading, isError, refetch } = useServices(projectId);

  return (
    <div>
      <PageHeader title="Services" description="Manage services, environment variables, and domains">
        <Button size="sm" onClick={() => setAddServiceOpen(true)}>
          <Plus className="mr-1.5 h-4 w-4" />
          Add service
        </Button>
      </PageHeader>

      {isLoading && <LoadingState rows={3} />}
      {isError && <ErrorState message="Could not load services." onRetry={() => refetch()} />}

      {!isLoading && !isError && data?.items.length === 0 && (
        <EmptyState
          title="No services"
          description="Add a service to manage its environment variables and domains."
        >
          <Button size="sm" onClick={() => setAddServiceOpen(true)}>
            <Plus className="mr-1.5 h-4 w-4" />
            Add service
          </Button>
        </EmptyState>
      )}

      {!isLoading && !isError && data && data.items.length > 0 && (
        <div className="p-6 space-y-6">
          {data.items.map((service) => (
            <ServiceCard key={service.id} projectId={projectId} service={service} />
          ))}
        </div>
      )}

      <AddServiceDialog
        projectId={projectId}
        open={addServiceOpen}
        onOpenChange={setAddServiceOpen}
      />
    </div>
  );
}

function ServiceCard({
  projectId,
  service,
}: {
  projectId: string;
  service: ProjectService;
}) {
  const [showEnv, setShowEnv] = useState(false);

  return (
    <div className="rounded-lg border">
      <div className="flex items-center justify-between p-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{service.name}</span>
            <Badge variant="secondary">:{service.containerPort}</Badge>
            {service.routingEnabled && (
              <Badge variant="outline">routing</Badge>
            )}
          </div>
          <p className="mt-0.5 text-xs text-muted-foreground">
            healthcheck: {service.healthcheckPath}:{service.healthcheckPort}
          </p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowEnv(!showEnv)}
        >
          {showEnv ? (
            <ChevronUp className="h-4 w-4" />
          ) : (
            <ChevronDown className="h-4 w-4" />
          )}
        </Button>
      </div>

      {showEnv && (
        <>
          <Separator />
          <div className="p-4 space-y-4">
            <EnvSection projectId={projectId} />
            <Separator />
            <DomainSection projectId={projectId} serviceId={service.id} />
          </div>
        </>
      )}
    </div>
  );
}

function EnvSection({ projectId }: { projectId: string }) {
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

      {activeSet && (
        <VariableList projectId={projectId} envSet={activeSet} />
      )}

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
                <Button variant="outline" onClick={() => setAddSetOpen(false)} type="button">
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

function VariableList({
  projectId,
  envSet,
}: {
  projectId: string;
  envSet: EnvironmentSet;
}) {
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
                  onClick={() =>
                    setRevealed((prev) => {
                      const next = new Set(prev);
                      next.has(v.id) ? next.delete(v.id) : next.add(v.id);
                      return next;
                    })
                  }
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
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <FormLabel className="!mt-0">Secret (masked in UI)</FormLabel>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" onClick={() => setAddOpen(false)} type="button">
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

function DomainSection({
  projectId,
  serviceId,
}: {
  projectId: string;
  serviceId: string;
}) {
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

  const domains = (data?.items ?? []).filter(
    (d) => d.projectServiceId === serviceId
  );

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
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <FormLabel className="!mt-0">Enable TLS</FormLabel>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button variant="outline" onClick={() => setAddOpen(false)} type="button">
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

function AddServiceDialog({
  projectId,
  open,
  onOpenChange,
}: {
  projectId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
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
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                  <FormLabel className="!mt-0">Enable routing via NPM</FormLabel>
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)} type="button">
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
