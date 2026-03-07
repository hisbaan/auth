"use client";

import * as React from "react";
import { UAParser } from "ua-parser-js";
import { Globe, Monitor, Smartphone, Tablet } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyableCode } from "@/components/ui/copyable-code";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { listAdminUserRefreshTokensAction } from "@/lib/actions";
import type { AdminUserSession } from "@/lib/session-utils";

type AdminUserSessionsProps = {
  sessions: AdminUserSession[];
  nextCursor?: string;
  userId: string;
};

const formatDateTime = (value?: string | null) => {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
};

const getStatusVariant = (status: AdminUserSession["status"]) => {
  if (status === "active") {
    return "default";
  }
  if (status === "expired") {
    return "secondary";
  }
  return "destructive";
};

type ClientInfo = {
  os: string;
  browser: string;
  deviceType: "mobile" | "tablet" | "desktop";
};

const parseClientInfo = (userAgent: string): ClientInfo => {
  if (!userAgent) {
    return { os: "Unknown OS", browser: "Unknown browser", deviceType: "desktop" };
  }

  const result = UAParser(userAgent);
  const osName = result.os.name ?? "Unknown OS";
  const browserName = result.browser.name ?? "Unknown browser";
  const deviceType = result.device.type === "mobile"
    ? "mobile"
    : result.device.type === "tablet"
      ? "tablet"
      : "desktop";

  return {
    os: osName,
    browser: browserName,
    deviceType,
  };
};

const deviceIconMap = {
  mobile: Smartphone,
  tablet: Tablet,
  desktop: Monitor,
};

export function AdminUserSessions({ sessions, nextCursor, userId }: AdminUserSessionsProps) {
  const [items, setItems] = React.useState(sessions);
  const [currentCursor, setCurrentCursor] = React.useState(nextCursor ?? null);
  const [isFirstPage, setIsFirstPage] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [isPending, startTransition] = React.useTransition();

  const loadPage = React.useCallback(
    (cursor?: string, resetToFirst = false) => {
      setError(null);
      startTransition(async () => {
        const result = await listAdminUserRefreshTokensAction(userId, cursor);
        if (result.error) {
          setError(result.error);
          return;
        }
        setItems(result.sessions);
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
    <div className="space-y-4">
      <div className="overflow-x-auto rounded-lg border border-border/60">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/30">
              <TableHead>Session</TableHead>
              <TableHead>Client</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Issued</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead>Revoked</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.length === 0 ? (
              <TableRow>
                <TableCell className="py-4 text-muted-foreground" colSpan={7}>
                  No sessions available yet.
                </TableCell>
              </TableRow>
            ) : (
              items.map((session) => {
                const client = parseClientInfo(session.userAgent);
                const DeviceIcon = deviceIconMap[client.deviceType];

                return (
                  <TableRow key={session.id}>
                    <TableCell>
                      <CopyableCode value={session.id} />
                    </TableCell>
                    <TableCell title={session.userAgent}>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-xs">
                          <DeviceIcon className="h-3 w-3 text-muted-foreground" />
                          <span className="font-medium text-foreground">
                            {client.os}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Globe className="h-3 w-3" />
                          <span>{client.browser}</span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono">
                      {session.ipAddress || "—"}
                    </TableCell>
                    <TableCell className="font-mono">
                      {formatDateTime(session.issuedAt)}
                    </TableCell>
                    <TableCell className="font-mono">
                      {formatDateTime(session.expiresAt)}
                    </TableCell>
                    <TableCell className="font-mono">
                      {formatDateTime(session.revokedAt)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusVariant(session.status)}>
                        {session.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center gap-2">
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
  );
}
