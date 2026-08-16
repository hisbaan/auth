import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { forgotPasswordAction } from "@/lib/actions";
import { withQuery } from "@/lib/http";

type ForgotPasswordPageProps = {
  searchParams: Promise<{
    next?: string;
  }>;
};

export default async function ForgotPasswordPage({ searchParams }: ForgotPasswordPageProps) {
  const params = await searchParams;

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Forgot password</CardTitle>
          <CardDescription>Send a password reset email to your account.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <form action={forgotPasswordAction} className="space-y-4">
            <input type="hidden" name="next" value={params.next ?? ""} />

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" name="email" type="email" required autoComplete="email" />
            </div>
            <Button className="w-full" type="submit">
              Send reset email
            </Button>
          </form>

          <Link href={withQuery("/login", { next: params.next })} className="text-sm text-muted-foreground hover:text-foreground">
            Back to sign in
          </Link>
        </CardContent>
      </Card>
    </main>
  );
}
