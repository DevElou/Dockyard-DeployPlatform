"use client";

import { useState } from "react";
import {
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  AlertTriangle,
  XCircle,
  Info,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/format";
import { humanizePhase } from "./phase-label";
import type {
  OperationEvent,
  OperationLevel,
} from "@/lib/types/operation-event";

interface EventTimelineProps {
  events: OperationEvent[];
}

const LEVEL_ICON: Record<OperationLevel, typeof CheckCircle2> = {
  info: Info,
  warn: AlertTriangle,
  error: XCircle,
  success: CheckCircle2,
};

const LEVEL_TONE: Record<OperationLevel, string> = {
  info: "text-muted-foreground",
  warn: "text-amber-600 dark:text-amber-400",
  error: "text-destructive",
  success: "text-emerald-600 dark:text-emerald-400",
};

export function EventTimeline({ events }: EventTimelineProps) {
  if (events.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No events yet.</p>
    );
  }

  return (
    <ol className="space-y-3">
      {events.map((ev) => (
        <TimelineRow key={ev.id} event={ev} />
      ))}
    </ol>
  );
}

function TimelineRow({ event }: { event: OperationEvent }) {
  const [open, setOpen] = useState(false);
  const Icon = LEVEL_ICON[event.level] ?? Info;
  const tone = LEVEL_TONE[event.level] ?? LEVEL_TONE.info;
  const hasDetails = event.details && Object.keys(event.details).length > 0;

  return (
    <li className="flex gap-3">
      <div className="pt-0.5">
        <Icon className={cn("h-4 w-4 shrink-0", tone)} aria-hidden />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
          <span className="text-sm font-medium">
            {humanizePhase(event.phase)}
          </span>
          <span className="text-xs text-muted-foreground">
            {formatDate(event.createdAt)}
          </span>
        </div>
        <p className="text-sm text-muted-foreground break-words">
          {event.message}
        </p>
        {hasDetails && (
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="mt-1 inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          >
            {open ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
            {open ? "Hide details" : "Show details"}
          </button>
        )}
        {open && hasDetails && (
          <dl className="mt-2 rounded-md border bg-muted/40 p-2 text-xs font-mono space-y-1">
            {Object.entries(event.details ?? {}).map(([k, v]) => (
              <div key={k} className="flex gap-2">
                <dt className="text-muted-foreground shrink-0">{k}</dt>
                <dd className="break-all whitespace-pre-wrap">{v}</dd>
              </div>
            ))}
          </dl>
        )}
      </div>
    </li>
  );
}
