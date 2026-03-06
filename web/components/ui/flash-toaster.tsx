"use client";

import * as React from "react";
import { toast } from "sonner";

type FlashPayload = {
  type: "success" | "error";
  message: string;
};

const FLASH_COOKIE = "flash";

function getFlashCookie(): FlashPayload | null {
  if (typeof document === "undefined") {
    return null;
  }

  const match = document.cookie
    .split("; ")
    .find((cookie) => cookie.startsWith(`${FLASH_COOKIE}=`));

  if (!match) {
    return null;
  }

  const value = match.split("=").slice(1).join("=");
  if (!value) {
    return null;
  }

  try {
    const decoded = decodeURIComponent(value);
    const parsed = JSON.parse(decoded) as FlashPayload;
    if (!parsed?.message || !parsed?.type) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function clearFlashCookie() {
  if (typeof document === "undefined") {
    return;
  }

  document.cookie = `${FLASH_COOKIE}=; Max-Age=0; path=/`;
}

export function FlashToaster() {
  React.useEffect(() => {
    const payload = getFlashCookie();
    if (!payload) {
      return;
    }

    if (payload.type === "success") {
      toast.success(payload.message);
    } else {
      toast.error(payload.message);
    }

    clearFlashCookie();
  }, []);

  return null;
}
