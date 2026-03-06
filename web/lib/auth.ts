import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { flashRedirectUrl } from "@/lib/flash";
import { getCurrentUser } from "@/lib/sdk";
import type { User } from "@/lib/types";

type WithAuthOptions = {
  loginRedirect?: string;
  requireRoles?: string[];
  unauthorizedRedirect?: string;
  unauthorizedMessage?: string;
};

type WithAuthResult = {
  user: User;
  cookieHeader: string;
};

export async function withAuth(
  options: WithAuthOptions,
): Promise<WithAuthResult> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();
  const user = await getCurrentUser(cookieHeader);

  if (!user) {
    redirect(options.loginRedirect ?? "/login");
  }

  if (options.requireRoles?.length) {
    const hasRoles = options.requireRoles.every((role) =>
      user.roles.includes(role),
    );

    if (!hasRoles) {
      const destination = options.unauthorizedRedirect ?? "/";
      if (options.unauthorizedMessage) {
        redirect(
          flashRedirectUrl(
            "error",
            options.unauthorizedMessage,
            destination,
          ),
        );
      }
      redirect(destination);
    }
  }

  return { user, cookieHeader };
}
