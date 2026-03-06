import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { resetPasswordAction } from "@/lib/actions";

type ResetPasswordProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function ResetPasswordPage({ searchParams }: ResetPasswordProps) {
  const params = await searchParams;

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Reset password</CardTitle>
          <CardDescription>Set a new secure password</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <form action={resetPasswordAction} className="space-y-4">
              <Input id="token" name="token" defaultValue={params.token ?? ""} required hidden />

            <div className="space-y-2">
              <Label htmlFor="new_password">New password</Label>
              <Input id="new_password" name="new_password" type="password" required minLength={8} />
            </div>

            <Button className="w-full" type="submit">
              Reset password
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
