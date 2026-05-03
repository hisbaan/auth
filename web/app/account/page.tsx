import { ClientActions } from "@/components/account/client-actions";
import { CreateClientDialog } from "@/components/account/create-client-dialog";
import { DeleteAccountDialog } from "@/components/account/delete-account-dialog";
import { ProfileForm } from "@/components/account/profile-form";
import { PasswordStrengthField } from "@/components/auth/password-strength-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CopyableCode } from "@/components/ui/copyable-code";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  createClientAction,
  deleteClientAction,
  deleteAccountAction,
  revokeAuthorizationAction,
  revokeClientAction,
  updateClientAction,
  updatePasswordAction,
  updateProfileAction,
} from "@/lib/actions";
import { withAuth } from "@/lib/auth";
import { listCurrentUserAuthorizations, listCurrentUserClients } from "@/lib/sdk";
import { getOIDCScopeDetail } from "@/lib/oidc";

const formatDateTime = (value?: string | null) => {
  if (!value) {
    return "-";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
};

export default async function AccountPage() {
  const { user: me, cookieHeader } = await withAuth({
    loginRedirect: "/login?next=/account",
  });
  const clientsResponse = await listCurrentUserClients(cookieHeader);
  const clients = clientsResponse?.clients ?? [];
  const authorizationsResponse = await listCurrentUserAuthorizations(cookieHeader);
  const authorizations = authorizationsResponse?.authorizations ?? [];

  return (
    <div className="min-h-screen">
      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-10 sm:px-6 lg:grid-cols-2">
        <h1 className="text-3xl font-semibold tracking-tight lg:col-span-2">
          Account
        </h1>
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <ProfileForm
              email={me.email}
              roles={me.roles}
              updateProfileAction={updateProfileAction}
              username={me.username}
            />
          </CardContent>
        </Card>

        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Security</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <form action={updatePasswordAction} className="space-y-4">
              <div className="flex flex-col gap-2">
                <Label htmlFor="current_password">Current password</Label>
                <Input
                  id="current_password"
                  name="current_password"
                  type="password"
                  required
                />
              </div>
              <PasswordStrengthField
                id="new_password"
                name="new_password"
                label="New password"
                autoComplete="new-password"
              />
              <div className="flex flex-col gap-2">
                <Label htmlFor="new_password">Confirm New password</Label>
                <Input
                  id="confirm_new_password"
                  name="confirm_new_password"
                  type="password"
                  required
                  minLength={8}
                />
              </div>
              <Button type="submit">Change password</Button>
            </form>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-start justify-between gap-4">
              OIDC Clients
              <CreateClientDialog createClientAction={createClientAction} />
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="overflow-x-auto rounded-lg border border-border/60">
              <Table className="min-w-[760px]">
                <TableHeader>
                  <TableRow className="bg-muted/30">
                    <TableHead>Name</TableHead>
                    <TableHead>Client ID</TableHead>
                    <TableHead>Redirect URI</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="w-[56px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {clients.length === 0 ? (
                    <TableRow>
                      <TableCell
                        className="py-8 text-muted-foreground text-center"
                        colSpan={6}
                      >
                        No clients yet. Create one to start testing your OIDC
                        flow.
                      </TableCell>
                    </TableRow>
                  ) : (
                    clients.map((client) => (
                      <TableRow key={client.id}>
                        <TableCell className="font-medium">
                          {client.name}
                        </TableCell>
                        <TableCell>
                          <CopyableCode value={client.id} />
                        </TableCell>
                        <TableCell className="max-w-[280px] truncate font-mono text-xs">
                          {client.redirect_uri}
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatDateTime(client.created_at)}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={
                              client.revoked_at ? "destructive" : "default"
                            }
                          >
                            {client.revoked_at ? "revoked" : "active"}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <ClientActions
                            client={client}
                            updateClientAction={updateClientAction}
                            revokeClientAction={revokeClientAction}
                            deleteClientAction={deleteClientAction}
                          />
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Connected Apps</CardTitle>
            <CardDescription>
              Apps you have granted access to through your account.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="overflow-x-auto rounded-lg border border-border/60">
              <Table className="min-w-[760px]">
                <TableHeader>
                  <TableRow className="bg-muted/30">
                    <TableHead>App</TableHead>
                    <TableHead>Client ID</TableHead>
                    <TableHead>Access</TableHead>
                    <TableHead>Last authorized</TableHead>
                    <TableHead className="w-[120px] text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {authorizations.length === 0 ? (
                    <TableRow>
                      <TableCell className="py-8 text-center text-muted-foreground" colSpan={5}>
                        No connected apps yet.
                      </TableCell>
                    </TableRow>
                  ) : (
                    authorizations.map((authorization) => (
                      <TableRow key={authorization.client_id}>
                        <TableCell>
                          <div className="space-y-1">
                            <div className="font-medium">{authorization.name}</div>
                            <div className="max-w-[280px] truncate font-mono text-xs text-muted-foreground">
                              {authorization.redirect_uri}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <CopyableCode value={authorization.client_id} />
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-2">
                            {authorization.granted_scopes.map((scope) => (
                              <Badge key={scope} variant="secondary" title={getOIDCScopeDetail(scope).description}>
                                {getOIDCScopeDetail(scope).title}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-xs">
                          {formatDateTime(authorization.last_authorized_at)}
                        </TableCell>
                        <TableCell className="text-right">
                          <form action={revokeAuthorizationAction}>
                            <input type="hidden" name="client_id" value={authorization.client_id} />
                            <Button type="submit" variant="outline">Disconnect</Button>
                          </form>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Danger Zone</CardTitle>
            <CardDescription>
              Once you delete your account, there is no going back
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DeleteAccountDialog email={me.email} deleteAccountAction={deleteAccountAction} />
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
