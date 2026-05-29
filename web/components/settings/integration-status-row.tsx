import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface Meta {
  label: string;
  value: string;
}

interface IntegrationStatusRowProps {
  name: string;
  description: string;
  enabled: boolean;
  meta?: Meta[];
}

export function IntegrationStatusRow({
  name,
  description,
  enabled,
  meta,
}: IntegrationStatusRowProps) {
  return (
    <div className="flex items-start justify-between py-3.5 border-b last:border-0">
      <div className="space-y-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{name}</span>
          <Badge
            variant="secondary"
            className={cn(
              "text-xs",
              enabled
                ? "bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300"
                : "text-muted-foreground"
            )}
          >
            {enabled ? "Connected" : "Not configured"}
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground">{description}</p>
        {enabled && meta && meta.length > 0 && (
          <div className="mt-1 space-y-0.5">
            {meta.map(({ label, value }) => (
              <p key={label} className="text-xs text-muted-foreground font-mono">
                <span className="font-sans font-medium not-italic">{label}:</span>{" "}
                {value}
              </p>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
