import Link from "next/link";

import { PasswordStrengthField } from "@/components/auth/password-strength-field";
import { SocialButtons } from "@/components/auth/social-buttons";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { registerAction } from "@/lib/actions";
import { withQuery } from "@/lib/http";

type RegisterPageProps = {
  searchParams: Promise<{
    next?: string;
  }>;
};

export default async function RegisterPage({ searchParams }: RegisterPageProps) {
  const params = await searchParams;

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create account</CardTitle>
          <CardDescription>Register with username, email, and password.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <form action={registerAction} className="space-y-4">
            <input type="hidden" name="next" value={params.next ?? ""} />

            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input id="username" name="username" required minLength={3} maxLength={32} autoComplete="username" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" name="email" type="email" required autoComplete="email" />
            </div>

            <PasswordStrengthField id="password" name="password" label="Password" autoComplete="new-password" />

            <Button className="w-full" type="submit">
              Create account
            </Button>
          </form>

          <SocialButtons />

          <p className="text-sm text-muted-foreground">
            Already registered?{" "}
            <Link href={withQuery("/login", { next: params.next })} className="text-foreground underline underline-offset-4">
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
