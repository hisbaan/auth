import Link from "next/link";

import { SocialButtons } from "@/components/auth/social-buttons";
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
import { loginAction } from "@/lib/actions";
import { cookies } from "next/headers";
import { getCurrentUser } from "@/lib/sdk";
import { redirect } from "next/navigation";
import { sanitizeRedirectPathOrUrl } from "@/lib/utils";
import { isAllowedCallbackUrl } from "@/lib/callback";
import { API_BASE_URL } from "@/lib/config";
import { withQuery } from "@/lib/http";

type LoginPageProps = {
  searchParams: Promise<{
    next?: string;
    callback_url?: string;
    state?: string;
    email?: string;
  }>;
};

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();

  const me = await getCurrentUser(cookieHeader);
  if (me) {
    const next = params.next ?? "";
    const callbackUrl = params.callback_url ?? "";
    const state = params.state ?? "";

    let destination = sanitizeRedirectPathOrUrl(next, "/account", [new URL(API_BASE_URL).origin]);
    if (callbackUrl && isAllowedCallbackUrl(callbackUrl)) {
      destination = `/authorize?callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}`;
    }

    redirect(destination);
  }

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Sign in</CardTitle>
          <CardDescription>
            Authenticate with your email and password.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <form action={loginAction} className="space-y-4">
            <input type="hidden" name="next" value={params.next ?? ""} />
            <input
              type="hidden"
              name="callback_url"
              value={params.callback_url ?? ""}
            />
            <input type="hidden" name="state" value={params.state ?? ""} />

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                defaultValue={params.email ?? ""}
                required
                autoComplete="email"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                required
                autoComplete="current-password"
              />
            </div>

            <Button className="w-full" type="submit">
              Sign in
            </Button>
          </form>

          <div className="space-y-4">
            <p className="text-xs uppercase tracking-wide text-muted-foreground">
              or continue with
            </p>
            <SocialButtons />
          </div>

          <div className="flex items-center justify-between text-sm">
            <Link
              href={withQuery("/forgot-password", { next: params.next })}
              className="text-muted-foreground hover:text-foreground"
            >
              Forgot password?
            </Link>
            <Link
              href={withQuery("/register", { next: params.next })}
              className="text-muted-foreground hover:text-foreground"
            >
              Create account
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
