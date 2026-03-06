import { ALLOWED_CALLBACK_HOSTS } from "@/lib/config";

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

export function buildAuthorizeSuccessUrl(callbackUrl: string, state?: string | null) {
  const url = new URL(callbackUrl);
  url.searchParams.set("auth", "success");
  if (state) {
    url.searchParams.set("state", state);
  }
  return url.toString();
}
