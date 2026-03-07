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
import { AdminUserSessions } from "@/components/admin/user-sessions";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { withAuth } from "@/lib/auth";
import { mapAdminUserEvents } from "@/lib/event-utils";
import { mapAdminUserSessions } from "@/lib/session-utils";
import {
  getAdminUser,
  listAdminUserEvents,
  listAdminUserRefreshTokens,
} from "@/lib/sdk";
import { Check } from "lucide-react";

type AdminUserDetailPageProps = {
  params: Promise<{ userId: string }>;
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

  const [user, eventsResponse, refreshTokensResponse] = await Promise.all([
    getAdminUser(cookieHeader, userId),
    listAdminUserEvents(cookieHeader, userId),
    listAdminUserRefreshTokens(cookieHeader, userId),
  ]);
  const roles = user?.roles ?? [];
  const { events, nextCursor } = mapAdminUserEvents(eventsResponse);
  const { sessions, nextCursor: sessionsNextCursor } =
    mapAdminUserSessions(refreshTokensResponse);

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
                <CardTitle>Sessions</CardTitle>
              </CardHeader>
              <CardContent>
                <AdminUserSessions
                  sessions={sessions}
                  nextCursor={sessionsNextCursor}
                  userId={userId}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="events" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Audit events</CardTitle>
              </CardHeader>
              <CardContent>
                <AdminUserEvents
                  events={events}
                  nextCursor={nextCursor}
                  userId={userId}
                />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
