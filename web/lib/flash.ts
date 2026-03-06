import { cookies } from "next/headers";

import { COOKIE_SECURE } from "@/lib/config";

const FLASH_COOKIE = "flash";

type FlashType = "success" | "error";

function flashCookieOptions(maxAge: number) {
  return {
    httpOnly: false,
    secure: COOKIE_SECURE,
    sameSite: "lax" as const,
    path: "/",
    maxAge,
  };
}

export async function setFlash(type: FlashType, message: string) {
  const cookieStore = await cookies();
  const payload = JSON.stringify({ type, message });
  cookieStore.set(FLASH_COOKIE, payload, flashCookieOptions(60));
}

export function flashRedirectUrl(
  type: FlashType,
  message: string,
  redirectTo: string,
) {
  const params = new URLSearchParams({
    type,
    message,
    redirect: redirectTo,
  });
  return `/flash?${params.toString()}`;
}
