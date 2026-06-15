import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { verifyEmail } from "@/lib/sdk";
import Link from "next/link";

type VerifyEmailProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function VerifyEmailPage({
  searchParams,
}: VerifyEmailProps) {
  const params = await searchParams;
  const token = params.token?.trim() ?? "";
  const result = token ? await verifyEmail(token) : null;
  const isVerified = result?.ok ?? false;
  const continueUrl = result?.data?.continue_url;
  const loginHref = continueUrl ? `/login?next=${encodeURIComponent(continueUrl)}` : "/login";

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>{isVerified ? "Email verified" : "Verify email"}</CardTitle>
          <CardDescription>
            {isVerified
              ? "Your email address has been verified."
              : "We could not verify your email address from this link."}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {isVerified
              ? continueUrl
                ? "You can now sign in and continue authorizing the application."
                : "You can now sign in and continue using your account."
              : token
                ? "This verification link is invalid, expired, or has already been used."
                : "This verification link is missing a token. Please use the link from your email."}
          </p>

          <Button className="w-full" asChild>
            <Link className="text-black!" href={isVerified ? loginHref : "/"}>
              {isVerified ? "Sign in" : "Back home"}
            </Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}
