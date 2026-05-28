import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { BuildStatusBadge } from "@/components/status/build-status-badge";
import { formatRelative } from "@/lib/format";
import type { Release } from "@/lib/types/release";

interface ReleasesTableProps {
  items: Release[];
}

export function ReleasesTable({ items }: ReleasesTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Version</TableHead>
          <TableHead>Ref</TableHead>
          <TableHead>Commit</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Created</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {items.map((r) => (
          <TableRow key={r.id}>
            <TableCell className="font-mono text-sm">{r.version}</TableCell>
            <TableCell className="font-mono text-sm">{r.gitRef}</TableCell>
            <TableCell className="font-mono text-xs text-muted-foreground">
              {r.gitSha ? r.gitSha.slice(0, 7) : "—"}
            </TableCell>
            <TableCell>
              <BuildStatusBadge status={r.buildStatus} />
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {formatRelative(r.createdAt)}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
