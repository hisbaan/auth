import Link from "next/link";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { resendVerificationAction } from "@/lib/actions";
import { withQuery } from "@/lib/http";

type VerifyEmailSentPageProps = {
  searchParams: Promise<{
    email?: string;
    next?: string;
    callback_url?: string;
    state?: string;
    sent?: string;
  }>;
};

export default async function VerifyEmailSentPage({
  searchParams,
}: VerifyEmailSentPageProps) {
  const params = await searchParams;
  const email = params.email?.trim() ?? "";
  const next = params.next?.trim() ?? "";
  // A sign-in attempt inside the resend cooldown leaves the earlier link valid instead of
  // sending a new one, so say that rather than claiming an email is on its way.
  const alreadySent = params.sent === "0";
  const loginHref = withQuery("/login", {
    next,
    callback_url: params.callback_url,
    state: params.state,
    email,
  });

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Check your email</CardTitle>
          <CardDescription>
            {alreadySent
              ? "A verification link was already sent to you recently."
              : "We sent you a link to verify your email address."}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2 text-sm text-muted-foreground">
            <p>
              {email ? (
                <>
                  Open the link we sent to{" "}
                  <span className="font-medium text-foreground">{email}</span> to
                  activate your account.
                </>
              ) : (
                "Open the link we sent you to activate your account."
              )}
            </p>
            <p>
              The link expires in 24 hours. If you cannot find it, check your spam
              folder.
            </p>
          </div>

          {email ? (
            <form action={resendVerificationAction}>
              <input type="hidden" name="email" value={email} />
              <input type="hidden" name="next" value={next} />

              <Button className="w-full" type="submit" variant="outline">
                Resend email
              </Button>
            </form>
          ) : null}

          <p className="text-sm text-muted-foreground">
            Already verified?{" "}
            <Link
              href={loginHref}
              className="text-foreground underline underline-offset-4"
            >
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
