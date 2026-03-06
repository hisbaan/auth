import Link from "next/link";

import { CopyableCode } from "@/components/ui/copyable-code";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AdminUserEvents } from "@/components/admin/user-events";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { withAuth } from "@/lib/auth";
import { getAdminUser } from "@/lib/sdk";
import { Check } from "lucide-react";

type AdminUserDetailPageProps = {
  params: Promise<{ userId: string }>;
};

const mockSessions = [
  {
    id: "sess_01J1WQ7F2W9NFK9Z9C2H2G0Z9A",
    device: "Chrome on macOS",
    ipAddress: "203.0.113.42",
    location: "San Francisco, CA",
    status: "active",
    lastActiveAt: "2026-03-05T16:22:10Z",
    createdAt: "2026-03-03T09:14:32Z",
  },
  {
    id: "sess_01J1S9Z4YAM8JQW7FQYB6RX4QH",
    device: "Safari on iOS",
    ipAddress: "198.51.100.18",
    location: "New York, NY",
    status: "revoked",
    lastActiveAt: "2026-02-28T21:05:48Z",
    createdAt: "2026-02-25T18:40:12Z",
  },
];

const mockEvents = [
  {
    id: "evt_01J1Z1K0N9M4F1E5A2C9D8Y7Q1",
    type: "user.login.succeeded",
    createdAt: "2026-03-05T16:22:10Z",
    payload: {
      ip: "203.0.113.42",
      user_agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_3)",
      session_id: "sess_01J1WQ7F2W9NFK9Z9C2H2G0Z9A",
    },
  },
  {
    id: "evt_01J1Z1H4J3N5C2V7Q0X3B8Z9P4",
    type: "user.role.assigned",
    createdAt: "2026-03-02T11:48:52Z",
    payload: {
      role: "admin",
      assigned_by: "user_01J0FKV1D1X7B9C8N3G2H1Q0P9",
    },
  },
  {
    id: "evt_01J1Z1F0W8D2M9Z5H7Q2X4N6V1",
    type: "user.profile.updated",
    createdAt: "2026-02-26T08:15:21Z",
    payload: {
      fields: ["email"],
      previous_email: "old-email@example.com",
      new_email: "new-email@example.com",
    },
  },
];

const formatDateTime = (value?: string) => {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
};

export default async function AdminUserDetailPage({
  params,
}: AdminUserDetailPageProps) {
  const { userId } = await params;

  const { cookieHeader } = await withAuth({
    loginRedirect: `/login?next=/admin/users/${userId}`,
    requireRoles: ["admin"],
    unauthorizedMessage: "Admin access required",
    unauthorizedRedirect: "/",
  });

  const user = await getAdminUser(cookieHeader, userId);
  const roles = user?.roles ?? [];

  return (
    <div className="min-h-screen">
      <main className="mx-auto w-full max-w-6xl space-y-8 px-4 py-10 sm:px-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
            <Button variant="ghost" size="sm" asChild>
              <Link href="/admin">Back to users</Link>
            </Button>
            {user ? (
              <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-3xl font-semibold tracking-tight">
                    {user.username}
                  </h1>
                  <Badge variant={user.email_verified ? "default" : "outline"}>
                    {user.email_verified ? "verified" : "unverified"}
                  </Badge>
                  {roles.length === 0 ? (
                    <Badge variant="outline">No roles</Badge>
                  ) : (
                    roles.map((role) => (
                      <Badge key={role} variant="secondary">
                        {role}
                      </Badge>
                    ))
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                  <span>{user.email}</span>
                  <span>•</span>
                  <CopyableCode value={user.id} />
                </div>
              </div>
            ) : (
              <div className="space-y-2">
                <h1 className="text-3xl font-semibold tracking-tight">
                  User not found
                </h1>
                <p className="text-muted-foreground">
                  We could not load this user. It may have been removed or you
                  may not have access.
                </p>
              </div>
            )}
          </div>
          {user ? (
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm">
                Reset password
              </Button>
              {user.email_verified || (
                <Button variant="outline" size="sm">
                  Verify Email
                </Button>
              )}
            </div>
          ) : null}
        </div>

        <Tabs defaultValue="overview" className="space-y-3">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="sessions">Sessions</TabsTrigger>
            <TabsTrigger value="events">Events</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            {!user ? (
              <Card>
                <CardHeader>
                  <CardTitle>Unable to load user</CardTitle>
                  <CardDescription>
                    Check the user ID and try again.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button variant="outline" asChild>
                    <Link href="/admin">Return to users</Link>
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <div className="grid gap-6 lg:grid-cols-3">
                <Card className="lg:col-span-2">
                  <CardHeader>
                    <CardTitle>User details</CardTitle>
                  </CardHeader>
                  <CardContent className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Username
                      </p>
                      <p className="text-sm font-medium text-foreground">
                        {user?.username}
                      </p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Email
                      </p>
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium text-foreground">
                          {user?.email}
                        </p>
                        {user?.email_verified ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className="inline-flex h-5 w-5 items-center justify-center rounded-full border border-emerald-500/40 bg-emerald-500/10 text-emerald-500">
                                <Check className="h-3 w-3" />
                              </span>
                            </TooltipTrigger>
                            <TooltipContent>Email Verified</TooltipContent>
                          </Tooltip>
                        ) : null}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        User ID
                      </p>
                      <CopyableCode value={user?.id ?? ""} />
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Roles
                      </p>
                      <div className="flex flex-wrap gap-2">
                        {roles.length === 0 ? (
                          <Badge variant="outline">No roles</Badge>
                        ) : (
                          roles.map((role) => (
                            <Badge key={role} variant="secondary">
                              {role}
                            </Badge>
                          ))
                        )}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Created at
                      </p>
                      <p className="text-sm font-medium text-foreground">
                        {formatDateTime(user?.created_at)}
                      </p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Updated at
                      </p>
                      <p className="text-sm font-medium text-foreground">
                        {formatDateTime(user?.updated_at)}
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}
          </TabsContent>

          <TabsContent value="sessions" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Active sessions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto rounded-lg border border-border/60">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-muted/30">
                        <TableHead>Session</TableHead>
                        <TableHead>Device</TableHead>
                        <TableHead>IP</TableHead>
                        <TableHead>Location</TableHead>
                        <TableHead>Last active</TableHead>
                        <TableHead>Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {mockSessions.map((session) => (
                        <TableRow key={session.id}>
                          <TableCell>
                            <CopyableCode value={session.id} />
                          </TableCell>
                          <TableCell className="font-mono">
                            {session.device}
                          </TableCell>
                          <TableCell className="font-mono">
                            {session.ipAddress}
                          </TableCell>
                          <TableCell className="font-mono">
                            {session.location}
                          </TableCell>
                          <TableCell className="font-mono">
                            {formatDateTime(session.lastActiveAt)}
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant={
                                session.status === "active"
                                  ? "default"
                                  : "outline"
                              }
                            >
                              {session.status}
                            </Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="events" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Audit events</CardTitle>
              </CardHeader>
              <CardContent>
                <AdminUserEvents events={mockEvents} />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
