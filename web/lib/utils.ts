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
