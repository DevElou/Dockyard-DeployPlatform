"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { DeploymentStatusBadge } from "@/components/status/deployment-status-badge";
import { formatRelative } from "@/lib/format";
import type { Deployment } from "@/lib/types/deployment";

interface DeploymentsTableProps {
  items: Deployment[];
  onSelect?: (deployment: Deployment) => void;
}

export function DeploymentsTable({ items, onSelect }: DeploymentsTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Status</TableHead>
          <TableHead>Release</TableHead>
          <TableHead>Target</TableHead>
          <TableHead>Strategy</TableHead>
          <TableHead>Started</TableHead>
          <TableHead>Finished</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {items.map((d) => (
          <TableRow
            key={d.id}
            onClick={onSelect ? () => onSelect(d) : undefined}
            className={onSelect ? "cursor-pointer" : undefined}
          >
            <TableCell>
              <DeploymentStatusBadge status={d.status} />
            </TableCell>
            <TableCell className="font-mono text-xs">
              {d.releaseId.slice(0, 8)}
            </TableCell>
            <TableCell className="font-mono text-xs">
              {d.runtimeTargetId.slice(0, 8)}
            </TableCell>
            <TableCell className="text-sm">{d.strategy}</TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {d.startedAt ? formatRelative(d.startedAt) : "—"}
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {d.finishedAt ? formatRelative(d.finishedAt) : "—"}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
