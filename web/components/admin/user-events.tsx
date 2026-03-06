"use client";

import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type AdminUserEvent = {
  id: string;
  type: string;
  createdAt: string;
  payload: Record<string, unknown>;
};

type AdminUserEventsProps = {
  events: AdminUserEvent[];
};

const formatDateTime = (value?: string) => {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
};

export function AdminUserEvents({ events }: AdminUserEventsProps) {
  const [selectedId, setSelectedId] = React.useState(
    events[0]?.id ?? null,
  );

  const selectedEvent = events.find((event) => event.id === selectedId) ?? null;

  return (
    <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
      <div className="space-y-2">
        {events.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No events available yet.
          </p>
        ) : (
          events.map((event) => (
            <Button
              key={event.id}
              type="button"
              variant="ghost"
              className={cn(
                "h-auto w-full items-start justify-between gap-4 rounded-lg border border-border/60 px-3 py-3 text-left",
                selectedId === event.id &&
                  "border-primary/40 bg-primary/5 text-foreground",
              )}
              onClick={() => setSelectedId(event.id)}
            >
              <div className="space-y-1">
                <p className="text-xs uppercase text-muted-foreground">
                  {formatDateTime(event.createdAt)}
                </p>
                <p className="text-sm font-medium text-foreground">
                  {event.type}
                </p>
              </div>
              <Badge variant="outline">Event</Badge>
            </Button>
          ))
        )}
      </div>
      <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
        {selectedEvent ? (
          <pre className="max-h-[520px] overflow-auto whitespace-pre-wrap break-words rounded-md bg-background p-4 text-xs font-mono text-foreground">
            {JSON.stringify(selectedEvent.payload, null, 2)}
          </pre>
        ) : (
          <div className="flex h-full min-h-[220px] items-center justify-center text-sm text-muted-foreground">
            Select an event to view its payload.
          </div>
        )}
      </div>
    </div>
  );
}
