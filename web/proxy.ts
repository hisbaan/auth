import { NextRequest, NextResponse } from "next/server";

import { API_BASE_URL, COOKIE_DOMAIN, COOKIE_SECURE } from "@/lib/config";

type RefreshResponse = {
  access_token: string;
  expires_in: number;
  refresh_token: string;
};

const REFRESH_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 7;
const EXPIRY_SKEW_SECONDS = 30;

function decodeJwtPayload(token: string): { exp?: number } | null {
  const parts = token.split(".");
  if (parts.length < 2) {
    return null;
  }

  try {
    const normalized = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    const decoded = atob(padded);
    return JSON.parse(decoded) as { exp?: number };
  } catch {
    return null;
  }
}

function isAccessTokenExpired(token: string): boolean {
  const payload = decodeJwtPayload(token);
  if (!payload?.exp) {
    return true;
  }

  const nowSeconds = Math.floor(Date.now() / 1000);
  return payload.exp <= nowSeconds + EXPIRY_SKEW_SECONDS;
}

async function refreshAccessToken(refreshToken: string): Promise<RefreshResponse | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
        Cookie: `refresh_token=${refreshToken}`,
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    const data = (await response.json()) as RefreshResponse;
    if (!data?.access_token || !data?.refresh_token || !data?.expires_in) {
      return null;
    }

    return data;
  } catch {
    return null;
  }
}

function setAuthCookies(response: NextResponse, data: RefreshResponse) {
  const options = {
    httpOnly: true,
    secure: COOKIE_SECURE,
    sameSite: "lax" as const,
    path: "/",
    ...(COOKIE_DOMAIN ? { domain: COOKIE_DOMAIN } : {}),
  };

  response.cookies.set("access_token", data.access_token, {
    ...options,
    maxAge: data.expires_in,
  });

  response.cookies.set("refresh_token", data.refresh_token, {
    ...options,
    maxAge: REFRESH_COOKIE_MAX_AGE_SECONDS,
  });
}

function clearAuthCookies(response: NextResponse) {
  const options = {
    httpOnly: true,
    secure: COOKIE_SECURE,
    sameSite: "lax" as const,
    path: "/",
    maxAge: 0,
    ...(COOKIE_DOMAIN ? { domain: COOKIE_DOMAIN } : {}),
  };

  response.cookies.set("access_token", "", options);
  response.cookies.set("refresh_token", "", options);
}

export async function proxy(request: NextRequest) {
  const refreshToken = request.cookies.get("refresh_token")?.value;
  if (!refreshToken) {
    return NextResponse.next();
  }

  const accessToken = request.cookies.get("access_token")?.value;
  const shouldRefresh = !accessToken || isAccessTokenExpired(accessToken);
  if (!shouldRefresh) {
    return NextResponse.next();
  }

  const refreshed = await refreshAccessToken(refreshToken);
  const response = NextResponse.next();

  if (!refreshed) {
    clearAuthCookies(response);
    return response;
  }

  setAuthCookies(response, refreshed);
  return response;
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
