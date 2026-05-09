import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function sanitizeRelativePath(value: string, fallback = "/") {
  if (!value) {
    return fallback
  }

  if (
    value.startsWith("http://") ||
    value.startsWith("https://") ||
    value.startsWith("//") ||
    value.includes("\\")
  ) {
    return fallback
  }

  if (!value.startsWith("/") || value.includes("..")) {
    return fallback
  }

  return value
}

export function sanitizeRedirectPathOrUrl(value: string, fallback = "/", allowedOrigins: string[] = []) {
  const relativePath = sanitizeRelativePath(value, "")
  if (relativePath) {
    return relativePath
  }

  try {
    const url = new URL(value)
    if ((url.protocol === "http:" || url.protocol === "https:") && allowedOrigins.includes(url.origin)) {
      return url.toString()
    }
  } catch {
    return fallback
  }

  return fallback
}
