import { PasswordStrengthField } from "@/components/auth/password-strength-field";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
            <input id="token" name="token" defaultValue={params.token ?? ""} required hidden />

            <PasswordStrengthField id="new_password" name="new_password" label="New password" autoComplete="new-password" />

            <Button className="w-full" type="submit">
              Reset password
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
