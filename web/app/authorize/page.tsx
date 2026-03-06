import { redirect } from "next/navigation";
import Link from "next/link";

import { SiteHeader } from "@/components/layout/site-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { buildAuthorizeSuccessUrl, isAllowedCallbackUrl } from "@/lib/callback";
import { DOCS_URL } from "@/lib/config";
import { withAuth } from "@/lib/auth";
import { withQuery } from "@/lib/http";

type AuthorizePageProps = {
  searchParams: Promise<{ callback_url?: string; state?: string }>;
};

export default async function AuthorizePage({ searchParams }: AuthorizePageProps) {
  const params = await searchParams;
  const callbackUrl = params.callback_url;
  const state = params.state;

  if (!callbackUrl) {
    return (
      <div className="min-h-screen">
        <main className="mx-auto flex w-full max-w-4xl justify-center px-4 py-10 sm:px-6">
          <Card className="w-full max-w-xl">
            <CardHeader>
              <CardTitle>Start an authorize request</CardTitle>
              <CardDescription>
                Pass a callback URL to complete sign-in for your app.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Developer?{" "}
                <Link href={DOCS_URL} className="text-foreground underline underline-offset-4">
                  Get started
                </Link>
              </p>
            </CardContent>
          </Card>
        </main>
      </div>
    );
  }

  if (!isAllowedCallbackUrl(callbackUrl)) {
    return (
      <div className="min-h-screen">
        <SiteHeader />
        <main className="mx-auto flex w-full max-w-4xl justify-center px-4 py-10 sm:px-6">
          <Card className="w-full max-w-xl">
            <CardHeader>
              <CardTitle>Invalid callback URL</CardTitle>
              <CardDescription>
                This callback host is not allowed.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Developer?{" "}
                <Link href={DOCS_URL} className="text-foreground underline underline-offset-4">
                  Get started
                </Link>
              </p>
              <Button variant="outline" asChild>
                <Link href="/">Back home</Link>
              </Button>
            </CardContent>
          </Card>
        </main>
      </div>
    );
  }

  await withAuth({
    loginRedirect: withQuery("/login", { callback_url: callbackUrl, state }),
  });

  redirect(buildAuthorizeSuccessUrl(callbackUrl, state));
}
