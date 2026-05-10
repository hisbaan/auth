const rawHosts = process.env.ALLOWED_CALLBACK_HOSTS ?? "localhost,.localhost,hisbaan.com,.hisbaan.com";

export const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:3000";
export const DOCS_URL = process.env.AUTH_DOCS_URL ?? "/docs";
export const COOKIE_DOMAIN = process.env.COOKIE_DOMAIN;
export const COOKIE_SECURE = process.env.NODE_ENV === "production";

export const ALLOWED_CALLBACK_HOSTS = rawHosts
  .split(",")
  .map((host) => host.trim().toLowerCase())
  .filter(Boolean);
