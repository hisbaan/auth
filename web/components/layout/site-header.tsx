import Link from "next/link";
import { cookies } from "next/headers";

import { Button } from "@/components/ui/button";
import { logoutAction } from "@/lib/actions";
import { getCurrentUser } from "@/lib/sdk";

type SiteHeaderProps = {
  /**
   * "minimal" strips every navigation affordance (logo link, auth buttons) so
   * users mid auth/OIDC flow cannot wander out of it and lose the flow state
   * carried in the URL.
   */
  variant?: "full" | "minimal";
};

export async function SiteHeader({ variant = "full" }: SiteHeaderProps) {
  if (variant === "minimal") {
    return (
      <header className="border-b border-border/70 bg-background/70 backdrop-blur">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 py-4 sm:px-6">
          <span className="font-mono text-sm font-medium tracking-wide text-muted-foreground">
            auth.hisbaan.com
          </span>
        </div>
      </header>
    );
  }

  const cookieStore = await cookies();
  const me = await getCurrentUser(cookieStore.toString());

  return (
    <header className="border-b border-border/70 bg-background/70 backdrop-blur">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 py-4 sm:px-6">
        <Link href="/" className="font-mono text-sm font-medium tracking-wide text-muted-foreground">
          auth.hisbaan.com
        </Link>

        <nav className="flex items-center gap-2">
          {me ? (
            <>
              <Button variant="ghost" asChild>
                <Link href="/account">Account</Link>
              </Button>
              {me.roles.includes("admin") ? (
                <Button variant="ghost" asChild>
                  <Link href="/admin">Admin</Link>
                </Button>
              ) : null}

              <form action={logoutAction}>
                <Button type="submit" variant="outline">
                  Logout
                </Button>
              </form>
            </>
          ) : (
            <>
              <Button variant="ghost" asChild>
                <Link href="/login">Login</Link>
              </Button>
              <Button asChild>
                <Link className="text-black!" href="/register">Create account</Link>
              </Button>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
