import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  deleteAccountAction,
  updatePasswordAction,
  updateProfileAction,
} from "@/lib/actions";
import { withAuth } from "@/lib/auth";

export default async function AccountPage() {
  const { user: me } = await withAuth({
    loginRedirect: "/login?next=/account",
  });

  return (
    <div className="min-h-screen">
      <main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-10 sm:px-6 lg:grid-cols-2">
        <div className="space-y-4 lg:col-span-2">
          <h1 className="text-3xl font-semibold tracking-tight">Account</h1>
          <p className="text-muted-foreground">
            Manage profile details, password, and your active session.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <form action={updateProfileAction} className="space-y-4">
              <div className="flex flex-col gap-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  name="username"
                  defaultValue={me.username}
                  required
                />
              </div>
              {/* TODO checkmark in the email field if the email is verified */}
              <div className="flex flex-col gap-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  defaultValue={me.email}
                  required
                />
              </div>
              <div className="flex flex-wrap gap-2">
                {me.roles.map((role) => (
                  <Badge key={role} variant="secondary">
                    {role}
                  </Badge>
                ))}
              </div>
              <Button type="submit">Save Changes</Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Security</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* TODO some live form validation. tanstack form? */}
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
              <div className="flex flex-col gap-2">
                <Label htmlFor="new_password">New password</Label>
                <Input
                  id="new_password"
                  name="new_password"
                  type="password"
                  required
                  minLength={8}
                />
              </div>
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

        <Card>
          <CardHeader>
            <CardTitle>Danger Zone</CardTitle>
            <CardDescription>
              Once you delete your account, there is no going back
            </CardDescription>
          </CardHeader>
          <CardContent>
            {/* TODO confirmation here, make you retype your email */}
            <form action={deleteAccountAction}>
              <Button type="submit" variant="destructive">
                Delete account
              </Button>
            </form>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
