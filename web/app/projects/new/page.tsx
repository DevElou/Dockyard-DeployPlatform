"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { useCreateProject } from "@/lib/hooks/use-projects";
import {
  createProjectSchema,
  parseGitHubUrl,
  slugify,
  type CreateProjectValues,
} from "@/lib/validation/project";
import { handleMutationError } from "@/lib/api-error";
import { toast } from "sonner";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

export default function NewProjectPage() {
  const router = useRouter();
  const { mutate, isPending } = useCreateProject();

  const form = useForm<CreateProjectValues>({
    resolver: zodResolver(createProjectSchema),
    defaultValues: {
      name: "",
      repoUrl: "",
      defaultBranch: "main",
      rootDirectory: ".",
      dockerfilePath: "Dockerfile",
      buildContext: ".",
    },
  });

  function onSubmit(values: CreateProjectValues) {
    const parsed = parseGitHubUrl(values.repoUrl);
    if (!parsed) {
      form.setError("repoUrl", { message: "Could not parse GitHub URL" });
      return;
    }

    mutate(
      {
        slug: slugify(values.name),
        name: values.name,
        githubOwner: parsed.owner,
        githubRepo: parsed.repo,
        defaultBranch: values.defaultBranch,
        rootDirectory: values.rootDirectory,
        dockerfilePath: values.dockerfilePath,
        buildContext: values.buildContext,
      },
      {
        onSuccess: (project) => {
          toast.success("Project created");
          router.push(`/projects/${project.id}`);
        },
        onError: (err) => handleMutationError(err, form),
      }
    );
  }

  return (
    <div>
      <PageHeader title="New project">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/projects">
            <ArrowLeft className="mr-1.5 h-4 w-4" />
            Back
          </Link>
        </Button>
      </PageHeader>

      <div className="mx-auto max-w-lg p-6">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="my-app" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="repoUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>GitHub repository</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="https://github.com/owner/repo"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="defaultBranch"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Default branch</FormLabel>
                  <FormControl>
                    <Input placeholder="main" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="rounded-md border p-4 space-y-4">
              <p className="text-sm font-medium text-muted-foreground">
                Advanced
              </p>

              <FormField
                control={form.control}
                name="dockerfilePath"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Dockerfile path</FormLabel>
                    <FormControl>
                      <Input placeholder="Dockerfile" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="buildContext"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Build context</FormLabel>
                    <FormControl>
                      <Input placeholder="." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="rootDirectory"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Root directory</FormLabel>
                    <FormDescription>
                      Relative to the repository root
                    </FormDescription>
                    <FormControl>
                      <Input placeholder="." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" asChild>
                <Link href="/projects">Cancel</Link>
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Creating…" : "Create project"}
              </Button>
            </div>
          </form>
        </Form>
      </div>
    </div>
  );
}
