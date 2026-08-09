import { ALLOWED_CALLBACK_HOSTS, API_BASE_URL } from "@/lib/config";

function hostAllowed(hostname: string) {
  const lower = hostname.toLowerCase();

  return ALLOWED_CALLBACK_HOSTS.some((rule) => {
    if (rule.startsWith(".")) {
      return lower === rule.slice(1) || lower.endsWith(rule);
    }

    return lower === rule;
  });
}

export function isAllowedCallbackUrl(value: string | null | undefined) {
  if (!value) {
    return false;
  }

  try {
    const url = new URL(value);
    const localHost = url.hostname === "localhost" || url.hostname === "127.0.0.1";
    if (url.protocol !== "https:" && !(url.protocol === "http:" && localHost)) {
      return false;
    }

    return hostAllowed(url.hostname);
  } catch {
    return false;
  }
}

// The API only accepts a verification return_to that points at its own /authorize endpoint,
// so anything else (a relative next, another origin) is dropped rather than sent and rejected.
export function emailVerificationReturnTo(next: string | null | undefined) {
  if (!next) {
    return "";
  }

  try {
    const url = new URL(next);
    const apiOrigin = new URL(API_BASE_URL).origin;
    return url.origin === apiOrigin && url.pathname === "/authorize" ? url.toString() : "";
  } catch {
    return "";
  }
}

export function buildAuthorizeSuccessUrl(callbackUrl: string, state?: string | null) {
  const url = new URL(callbackUrl);
  url.searchParams.set("auth", "success");
  if (state) {
    url.searchParams.set("state", state);
  }
  return url.toString();
}
