"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import { listAdminUserEventsAction } from "@/lib/actions";
import type { AdminUserEvent } from "@/lib/event-utils";
import { cn } from "@/lib/utils";

type AdminUserEventsProps = {
  events: AdminUserEvent[];
  nextCursor?: string;
  userId: string;
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

export function AdminUserEvents({ events, nextCursor, userId }: AdminUserEventsProps) {
  const [items, setItems] = React.useState(events);
  const [currentCursor, setCurrentCursor] = React.useState(nextCursor ?? null);
  const [selectedId, setSelectedId] = React.useState(items[0]?.id ?? null);
  const [isFirstPage, setIsFirstPage] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [isPending, startTransition] = React.useTransition();

  React.useEffect(() => {
    setSelectedId(items[0]?.id ?? null);
  }, [items]);

  const selectedEvent = items.find((event) => event.id === selectedId) ?? null;

  const loadPage = React.useCallback(
    (cursor?: string, resetToFirst = false) => {
      setError(null);
      startTransition(async () => {
        const result = await listAdminUserEventsAction(userId, cursor);
        if (result.error) {
          setError(result.error);
          return;
        }
        setItems(result.events);
        setCurrentCursor(result.nextCursor ?? null);
        setIsFirstPage(resetToFirst);
      });
    },
    [startTransition, userId],
  );

  const handleNext = () => {
    if (!currentCursor || isPending) {
      return;
    }
    loadPage(currentCursor, false);
  };

  const handleFirst = () => {
    if (isFirstPage || isPending) {
      return;
    }
    loadPage(undefined, true);
  };

  return (
    <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
      <div className="space-y-2">
        {items.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No events available yet.
          </p>
        ) : (
          items.map((event) => (
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
            </Button>
          ))
        )}
        <div className="flex items-center gap-2 pt-1">
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={isPending || isFirstPage}
            onClick={handleFirst}
          >
            First
          </Button>
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={isPending || !currentCursor}
            onClick={handleNext}
          >
            Next
          </Button>
        </div>
        {error ? (
          <p className="text-xs text-destructive">{error}</p>
        ) : null}
      </div>
      <div className="rounded-lg border border-border/60 bg-muted/30 p-4">
        {selectedEvent ? (
          <div className="space-y-3">
            <div className="rounded-md border border-border/60 bg-background p-3">
              <div className="grid gap-3 text-xs sm:grid-cols-2">
                <div className="space-y-1">
                  <p className="text-[10px] uppercase text-muted-foreground">
                    Type
                  </p>
                  <p className="text-xs font-medium text-foreground">
                    {selectedEvent.type}
                  </p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] uppercase text-muted-foreground">
                    Occurred
                  </p>
                  <p className="text-xs font-medium text-foreground">
                    {formatDateTime(selectedEvent.createdAt)}
                  </p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] uppercase text-muted-foreground">
                    IP address
                  </p>
                  <p className="text-xs font-mono text-foreground">
                    {selectedEvent.ipAddress || "—"}
                  </p>
                </div>
                <div className="space-y-1">
                  <p className="text-[10px] uppercase text-muted-foreground">
                    User agent
                  </p>
                  <p className="text-xs font-mono break-words text-foreground">
                    {selectedEvent.userAgent || "—"}
                  </p>
                </div>
              </div>
            </div>
            {selectedEvent.payload ? (
              <pre className="max-h-[520px] overflow-auto whitespace-pre-wrap break-words rounded-md bg-background p-4 text-xs font-mono text-foreground">
                {JSON.stringify(selectedEvent.payload, null, 2)}
              </pre>
            ) : (
              <div className="flex h-full min-h-[220px] items-center justify-center rounded-md border border-border/60 bg-background text-xs text-muted-foreground">
                No payload data for this event.
              </div>
            )}
          </div>
        ) : (
          <div className="flex h-full min-h-[220px] items-center justify-center text-sm text-muted-foreground">
            Select an event to view its payload.
          </div>
        )}
      </div>
    </div>
  );
}
