import { headers as nextHeaders } from "next/headers";

import { API_BASE_URL } from "@/lib/config";
import type {
	Client,
	AuthorizeClientInfo,
	ListEventsResponse,
	ListClientAuthorizationsResponse,
	ListClientsResponse,
	ListRefreshTokensResponse,
	ListRolesResponse,
	ListUsersResponse,
	LoginResponse,
	User,
} from "@/lib/types";

type SDKRequestOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: unknown;
  cookieHeader?: string;
  headers?: Record<string, string>;
};

async function forwardedClientHeaders() {
  const incoming = await nextHeaders();
  const headers: Record<string, string> = {};
  const forwardedFor = incoming.get("x-forwarded-for");
  const realIp = incoming.get("x-real-ip");
  const userAgent = incoming.get("user-agent");

  if (forwardedFor) {
    headers["X-Forwarded-For"] = forwardedFor;
  }
  if (realIp) {
    headers["X-Real-IP"] = realIp;
  }
  if (userAgent) {
    headers["User-Agent"] = userAgent;
  }

  return headers;
}

export type SDKResult<T> = {
  ok: boolean;
  status: number;
  data?: T;
  error?: string;
  headers: Headers;
};

export async function sdkRequest<T>(path: string, options: SDKRequestOptions = {}): Promise<SDKResult<T>> {
  const headers = new Headers();
  headers.set("Accept", "application/json");

  for (const [key, value] of Object.entries(await forwardedClientHeaders())) {
    headers.set(key, value);
  }

  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  if (options.cookieHeader) {
    headers.set("Cookie", options.cookieHeader);
  }

  if (options.headers) {
    for (const [key, value] of Object.entries(options.headers)) {
      if (value) {
        headers.set(key, value);
      }
    }
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
    cache: "no-store",
  });

  const contentType = response.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json");

  if (!response.ok) {
    const error = isJson ? JSON.stringify(await response.json()) : (await response.text()) || "Request failed";
    return {
      ok: false,
      status: response.status,
      error,
      headers: response.headers,
    };
  }

  const data = isJson ? ((await response.json()) as T) : undefined;
  return {
    ok: true,
    status: response.status,
    data,
    headers: response.headers,
  };
}

export async function loginWithPassword(email: string, password: string) {
  return sdkRequest<LoginResponse>("/auth/login", {
    method: "POST",
    body: { email, password },
  });
}

export async function registerUser(username: string, email: string, password: string) {
  return sdkRequest<void>("/auth/register", {
    method: "POST",
    body: { username, email, password },
  });
}

export async function sendForgotPassword(email: string) {
  return sdkRequest<void>("/auth/forgot-password", {
    method: "POST",
    body: { email },
  });
}

export async function resetPassword(token: string, newPassword: string) {
  return sdkRequest<void>("/auth/password-reset", {
    method: "POST",
    body: { token, new_password: newPassword },
  });
}

export async function verifyEmail(token: string) {
  return sdkRequest<void>("/auth/verify-email", {
    method: "POST",
    body: { token },
  });
}

export async function logout(cookieHeader: string) {
  return sdkRequest<void>("/auth/logout", {
    method: "POST",
    cookieHeader,
  });
}

// look at how withAuth works in other places like workos, this should be built into the next SDK
// export async function withAuth() {
//   const cookieStore = await cookies();
//   const accessToken = cookieStore.get("access_token");
//
//   if (!accessToken) {
//     return undefined;
//   }
//
//   // TODO parse access token, if expired get refresh? maybe this should be called in middleware, take a look at
// }

export async function getCurrentUser(cookieHeader: string): Promise<User | null> {
  const result = await sdkRequest<User>("/users/me", { cookieHeader });
  if (!result.ok || !result.data) {
    return null;
  }

  return result.data;
}

export async function updateCurrentUser(cookieHeader: string, username: string, email: string) {
  return sdkRequest<void>("/users/me", {
    method: "PUT",
    cookieHeader,
    body: { username, email },
  });
}

export async function updateCurrentPassword(cookieHeader: string, currentPassword: string, newPassword: string) {
  return sdkRequest<void>("/users/me/password", {
    method: "POST",
    cookieHeader,
    body: {
      current_password: currentPassword,
      new_password: newPassword,
    },
  });
}

export async function deleteCurrentUser(cookieHeader: string) {
  return sdkRequest<void>("/users/me", {
    method: "DELETE",
    cookieHeader,
  });
}

type ListCurrentUserClientsOptions = {
	limit?: number;
	cursor?: string;
};

export async function listCurrentUserClients(
	cookieHeader: string,
	options: ListCurrentUserClientsOptions = {},
): Promise<ListClientsResponse | null> {
	const params = new URLSearchParams();
	if (options.limit) {
		params.set("limit", String(options.limit));
	}
	if (options.cursor) {
		params.set("cursor", options.cursor);
	}

	const query = params.toString();
	const path = query ? `/users/me/clients?${query}` : "/users/me/clients";
	const result = await sdkRequest<ListClientsResponse>(path, { cookieHeader });
	if (!result.ok || !result.data) {
		return null;
	}

	return result.data;
}

export async function listCurrentUserAuthorizations(
	cookieHeader: string,
): Promise<ListClientAuthorizationsResponse | null> {
	const result = await sdkRequest<ListClientAuthorizationsResponse>("/users/me/authorizations", {
		cookieHeader,
	});
	if (!result.ok || !result.data) {
		return null;
	}

	return result.data;
}

type CreateCurrentUserClientParams = {
	name: string;
	redirectURI: string;
	allowedScopes: string[];
};

type UpdateCurrentUserClientParams = {
	name: string;
	redirectURI: string;
	allowedScopes: string[];
};

export async function createCurrentUserClient(
	cookieHeader: string,
	params: CreateCurrentUserClientParams,
) {
	return sdkRequest<Client>("/users/me/clients", {
		method: "POST",
		cookieHeader,
		body: {
			name: params.name,
			redirect_uri: params.redirectURI,
			allowed_scopes: params.allowedScopes,
		},
	});
}

export async function revokeCurrentUserClient(cookieHeader: string, clientId: string) {
	return sdkRequest<void>(`/users/me/clients/${encodeURIComponent(clientId)}/revoke`, {
		method: "POST",
		cookieHeader,
	});
}

export async function revokeCurrentUserAuthorization(cookieHeader: string, clientId: string) {
	return sdkRequest<void>(`/users/me/authorizations/${encodeURIComponent(clientId)}/revoke`, {
		method: "POST",
		cookieHeader,
	});
}

export async function updateCurrentUserClient(
	cookieHeader: string,
	clientId: string,
	params: UpdateCurrentUserClientParams,
) {
	return sdkRequest<void>(`/users/me/clients/${encodeURIComponent(clientId)}`, {
		method: "PUT",
		cookieHeader,
		body: {
			name: params.name,
			redirect_uri: params.redirectURI,
			allowed_scopes: params.allowedScopes,
		},
	});
}

export async function deleteCurrentUserClient(cookieHeader: string, clientId: string) {
	return sdkRequest<void>(`/users/me/clients/${encodeURIComponent(clientId)}`, {
		method: "DELETE",
		cookieHeader,
	});
}

export async function getAuthorizeClientInfo(cookieHeader: string, clientId: string): Promise<AuthorizeClientInfo | null> {
	const result = await sdkRequest<AuthorizeClientInfo>(`/authorize/client-info?client_id=${encodeURIComponent(clientId)}`, {
		cookieHeader,
	});
	if (!result.ok || !result.data) {
		return null;
	}

	return result.data;
}

export async function grantAuthorizeConsent(cookieHeader: string, clientId: string, scope: string) {
	return sdkRequest<void>("/authorize/consent", {
		method: "POST",
		cookieHeader,
		body: {
			client_id: clientId,
			scope,
		},
	});
}

export async function getRoles(cookieHeader: string): Promise<string[]> {
  const result = await sdkRequest<ListRolesResponse>("/roles", { cookieHeader });
  if (!result.ok || !result.data) {
    return [];
  }

  return result.data.roles ?? [];
}

export async function listAdminUsers(cookieHeader: string, limit = 50): Promise<ListUsersResponse | null> {
  const result = await sdkRequest<ListUsersResponse>(`/admin/users?limit=${limit}`, { cookieHeader });
  if (!result.ok || !result.data) {
    return null;
  }

  return result.data;
}

export async function getAdminUser(cookieHeader: string, userId: string): Promise<User | null> {
  const result = await sdkRequest<User>(`/admin/users/${encodeURIComponent(userId)}`, {
    cookieHeader,
  });
  if (!result.ok || !result.data) {
    return null;
  }

  return result.data;
}

type ListAdminUserEventsOptions = {
  limit?: number;
  cursor?: string;
};

export async function listAdminUserEvents(
  cookieHeader: string,
  userId: string,
  options: ListAdminUserEventsOptions = {},
): Promise<ListEventsResponse | null> {
  const params = new URLSearchParams();
  if (options.limit) {
    params.set("limit", String(options.limit));
  }
  if (options.cursor) {
    params.set("cursor", options.cursor);
  }

  const query = params.toString();
  const path = query
    ? `/admin/events/users/${encodeURIComponent(userId)}?${query}`
    : `/admin/events/users/${encodeURIComponent(userId)}`;
  const result = await sdkRequest<ListEventsResponse>(path, { cookieHeader });
  if (!result.ok || !result.data) {
    return null;
  }

	return result.data;
}

type ListAdminUserRefreshTokensOptions = {
	limit?: number;
	cursor?: string;
};

export async function listAdminUserRefreshTokens(
	cookieHeader: string,
	userId: string,
	options: ListAdminUserRefreshTokensOptions = {},
): Promise<ListRefreshTokensResponse | null> {
	const params = new URLSearchParams();
	if (options.limit) {
		params.set("limit", String(options.limit));
	}
	if (options.cursor) {
		params.set("cursor", options.cursor);
	}

	const query = params.toString();
	const path = query
		? `/admin/refresh-tokens/users/${encodeURIComponent(userId)}?${query}`
		: `/admin/refresh-tokens/users/${encodeURIComponent(userId)}`;
	const result = await sdkRequest<ListRefreshTokensResponse>(path, {
		cookieHeader,
	});
	if (!result.ok || !result.data) {
		return null;
	}

	return result.data;
}

export async function createRole(cookieHeader: string, name: string) {
  return sdkRequest<void>("/admin/roles", {
    method: "POST",
    cookieHeader,
    body: { name },
  });
}

export async function deleteRole(cookieHeader: string, name: string) {
  return sdkRequest<void>(`/admin/roles/${encodeURIComponent(name)}`, {
    method: "DELETE",
    cookieHeader,
  });
}

export async function addUserRole(cookieHeader: string, userId: string, role: string) {
  return sdkRequest<void>(`/admin/users/${encodeURIComponent(userId)}/roles`, {
    method: "POST",
    cookieHeader,
    body: {
      role,
    },
  });
}

export async function removeUserRole(cookieHeader: string, userId: string, role: string) {
  return sdkRequest<void>(`/admin/users/${encodeURIComponent(userId)}/roles/${encodeURIComponent(role)}`, {
    method: "DELETE",
    cookieHeader,
  });
}
