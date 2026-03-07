import type { ListRefreshTokensResponse, RefreshTokenResponse } from "@/lib/types";

export type AdminUserSession = {
  id: string;
  issuedAt: string;
  expiresAt: string;
  revokedAt?: string | null;
  ipAddress: string;
  userAgent: string;
  status: "active" | "expired" | "revoked";
};

export type AdminUserSessionsPage = {
  sessions: AdminUserSession[];
  nextCursor?: string;
};

function mapAdminUserSession(token: RefreshTokenResponse): AdminUserSession {
  return {
    id: token.id,
    issuedAt: token.issued_at,
    expiresAt: token.expires_at,
    revokedAt: token.revoked_at ?? null,
    ipAddress: token.ip_address,
    userAgent: token.user_agent,
    status: token.status,
  };
}

export function mapAdminUserSessions(
  response: ListRefreshTokensResponse | null | undefined,
): AdminUserSessionsPage {
  if (!response) {
    return { sessions: [] };
  }

  return {
    sessions: response.refresh_tokens.map(mapAdminUserSession),
    nextCursor: response.next_cursor,
  };
}
