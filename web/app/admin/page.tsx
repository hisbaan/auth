import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { AdminUserRow } from "@/components/admin/user-row";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  createRoleAction,
  deleteRoleAction,
} from "@/lib/actions";
import { withAuth } from "@/lib/auth";
import { getRoles, listAdminUsers } from "@/lib/sdk";

export default async function AdminPage() {
  const { cookieHeader } = await withAuth({
    loginRedirect: "/login?next=/admin",
    requireRoles: ["admin"],
    unauthorizedMessage: "Admin access required",
    unauthorizedRedirect: "/",
  });

  const [users, roles] = await Promise.all([
    listAdminUsers(cookieHeader),
    getRoles(cookieHeader),
  ]);

  return (
    <div className="min-h-screen">
      <main className="mx-auto w-full max-w-6xl space-y-6 px-4 py-10 sm:px-6">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Admin users</h1>
          <p className="text-muted-foreground">Manage users and roles.</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex w-full items-center justify-between">
              Roles
              <Dialog>
                <DialogTrigger asChild>
                  <Button size="sm" variant="secondary">
                    Add role
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Add role</DialogTitle>
                    <DialogDescription>
                      Create a new role for assigning to users.
                    </DialogDescription>
                  </DialogHeader>
                  <form action={createRoleAction}>
                    <div className="space-y-4">
                      <Input name="name" placeholder="editor" required />
                      <DialogFooter>
                        <Button type="submit">Create role</Button>
                      </DialogFooter>
                    </div>
                  </form>
                </DialogContent>
              </Dialog>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="overflow-x-auto rounded-lg border border-border/60">
              <Table className="min-w-[360px]">
                <TableHeader>
                  <TableRow className="bg-muted/30">
                    <TableHead>Role</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {roles.length === 0 ? (
                    <TableRow>
                      <TableCell
                        className="py-4 text-muted-foreground"
                        colSpan={3}
                      >
                        No roles found.
                      </TableCell>
                    </TableRow>
                  ) : (
                    roles.map((role) => (
                      <TableRow key={role}>
                        <TableCell className="font-mono">{role}</TableCell>
                        <TableCell className="justify-self-end">
                          <details className="inline-block">
                            <summary className="list-none cursor-pointer rounded-md border border-destructive/70 bg-destructive/10 px-2.5 py-1.5 text-xs font-medium text-destructive hover:bg-destructive/20">
                              Delete
                            </summary>
                            <div className="mt-2 rounded-md border border-border/60 bg-background p-3">
                              <p className="mb-2 text-xs text-muted-foreground">
                                Confirm delete for role: {role}
                              </p>
                              <form action={deleteRoleAction}>
                                <input type="hidden" name="name" value={role} />
                                <Button
                                  type="submit"
                                  variant="destructive"
                                  size="sm"
                                >
                                  Confirm
                                </Button>
                              </form>
                            </div>
                          </details>
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
            <CardTitle>Users</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto rounded-lg border border-border/60">
              <Table className="min-w-[360px]">
                <TableHeader>
                  <TableRow className="bg-muted/30">
                    <TableHead>Username</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>User ID</TableHead>
                    <TableHead>Roles</TableHead>
                    <TableHead>Verified</TableHead>
                    <TableHead></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users?.users.map((user) => (
                    <AdminUserRow
                      key={user.id}
                      user={user}
                      roles={roles}
                    />
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
