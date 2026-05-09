"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { buildAuthorizeSuccessUrl, isAllowedCallbackUrl } from "@/lib/callback";
import { API_BASE_URL, COOKIE_DOMAIN, COOKIE_SECURE } from "@/lib/config";
import { parseEncodedRequest, withEncodedRequest } from "@/lib/http";
import {
	addUserRole,
	createCurrentUserClient,
	createRole,
	deleteCurrentUserClient,
	deleteCurrentUser,
	deleteRole,
	grantAuthorizeConsent,
	listAdminUserEvents,
	listAdminUserRefreshTokens,
	listCurrentUserClients,
	loginWithPassword,
	logout,
	revokeCurrentUserAuthorization,
	registerUser,
	revokeCurrentUserClient,
	removeUserRole,
	resetPassword,
	sendForgotPassword,
	updateCurrentUserClient,
	updateCurrentPassword,
	updateCurrentUser,
} from "@/lib/sdk";
import { mapAdminUserEvents } from "@/lib/event-utils";
import { mapAdminUserSessions } from "@/lib/session-utils";
import { setFlash } from "@/lib/flash";
import { sanitizeRedirectPathOrUrl } from "@/lib/utils";

function cookieOptions(maxAge: number) {
  return {
    httpOnly: true,
    secure: COOKIE_SECURE,
    sameSite: "lax" as const,
    path: "/",
    maxAge,
    ...(COOKIE_DOMAIN ? { domain: COOKIE_DOMAIN } : {}),
  };
}

async function clearAuthCookies() {
  const cookieStore = await cookies();
  cookieStore.set("access_token", "", cookieOptions(0));
  cookieStore.set("refresh_token", "", cookieOptions(0));
}

export async function loginAction(formData: FormData) {
  const email = String(formData.get("email") ?? "").trim();
  const password = String(formData.get("password") ?? "");
  const next = String(formData.get("next") ?? "").trim();
  const callbackUrl = String(formData.get("callback_url") ?? "").trim();
  const state = String(formData.get("state") ?? "").trim();
  const loginRedirect = `/login?next=${encodeURIComponent(next)}&callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}&email=${encodeURIComponent(email)}`;

  if (!email || !password) {
    await setFlash("error", "Invalid credentials");
    redirect(loginRedirect);
  }

  const result = await loginWithPassword(email, password);
  if (!result.ok || !result.data) {
    await setFlash("error", "Invalid credentials");
    redirect(loginRedirect);
  }

  const cookieStore = await cookies();
  cookieStore.set("access_token", result.data.access_token, cookieOptions(result.data.expires_in));
  cookieStore.set("refresh_token", result.data.refresh_token, cookieOptions(60 * 60 * 24 * 7));

  let destination = sanitizeRedirectPathOrUrl(next, "/account", [new URL(API_BASE_URL).origin]);
  if (callbackUrl && isAllowedCallbackUrl(callbackUrl)) {
    destination = `/authorize?callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}`;
  }

  redirect(destination);
}

export async function registerAction(formData: FormData) {
  const username = String(formData.get("username") ?? "").trim();
  const email = String(formData.get("email") ?? "").trim();
  const password = String(formData.get("password") ?? "");

  if (!username || !email || !password) {
    await setFlash("error", "All fields are required");
    redirect("/register");
  }

  const result = await registerUser(username, email, password);
  if (!result.ok) {
    await setFlash("error", "Unable to create account");
    redirect("/register");
  }

  await setFlash("success", "Account created. Check your email to verify.");
  redirect("/register");
}

export async function forgotPasswordAction(formData: FormData) {
  const email = String(formData.get("email") ?? "").trim();
  if (!email) {
    await setFlash("error", "Email is required");
    redirect("/forgot-password");
  }

  const result = await sendForgotPassword(email);
  if (!result.ok) {
    await setFlash("error", "Unable to send reset email");
    redirect("/forgot-password");
  }

  await setFlash("success", "If this email exists, a reset link was sent.");
  redirect("/forgot-password");
}

export async function resetPasswordAction(formData: FormData) {
  const token = String(formData.get("token") ?? "").trim();
  const newPassword = String(formData.get("new_password") ?? "").trim();

  if (!token || !newPassword) {
    await setFlash("error", "Token and password are required");
    redirect(`/reset-password?token=${encodeURIComponent(token)}`);
  }

  const result = await resetPassword(token, newPassword);
  if (!result.ok) {
    await setFlash("error", "Unable to reset password");
    redirect(`/reset-password?token=${encodeURIComponent(token)}`);
  }

  await setFlash("success", "Password reset complete. You can sign in now.");
  redirect("/login");
}

export async function logoutAction() {
  const cookieStore = await cookies();
  await logout(cookieStore.toString());
  await clearAuthCookies();
  redirect("/");
}

export async function authorizeConsentAction(formData: FormData) {
	const request = String(formData.get("request") ?? "").trim();
	const authorizeParams = parseEncodedRequest(request);
	const clientId = authorizeParams?.get("client_id")?.trim() ?? "";
	const scope = authorizeParams?.get("scope")?.trim() ?? "";

	if (!request || !clientId || !scope) {
		await setFlash("error", "Invalid authorize request");
		redirect("/");
	}

	const cookieStore = await cookies();
	const result = await grantAuthorizeConsent(cookieStore.toString(), clientId, scope);
	if (!result.ok) {
		await setFlash("error", "Unable to grant consent");
		redirect(withEncodedRequest("/authorize", request));
	}

	redirect(`${API_BASE_URL}/authorize?${request}`);
}

export async function updateProfileAction(formData: FormData) {
  const username = String(formData.get("username") ?? "").trim();
  const email = String(formData.get("email") ?? "").trim();

  if (!username || !email) {
    await setFlash("error", "Username and email are required");
    redirect("/account");
  }

	const cookieStore = await cookies();
	const result = await updateCurrentUser(cookieStore.toString(), username, email);
	if (!result.ok || !result.data) {
		await setFlash("error", "Unable to update profile");
		redirect("/account");
	}

	if (result.data.email_verification_required) {
		await setFlash("success", "Profile updated. Check your email to verify the new address.");
		redirect("/account");
	}

	await setFlash("success", "Profile updated");
	redirect("/account");
}

export async function updatePasswordAction(formData: FormData) {
  const currentPassword = String(formData.get("current_password") ?? "");
  const newPassword = String(formData.get("new_password") ?? "");

  if (!currentPassword || !newPassword) {
    await setFlash("error", "Current and new password are required");
    redirect("/account");
  }

  const cookieStore = await cookies();
  const result = await updateCurrentPassword(cookieStore.toString(), currentPassword, newPassword);
  if (!result.ok) {
    await setFlash("error", "Unable to change password");
    redirect("/account");
  }

  await setFlash("success", "Password updated");
  redirect("/account");
}

export async function deleteAccountAction(formData: FormData) {
  const email = String(formData.get("email") ?? "");
  if (!email) {
    await setFlash("error", "Enter your email to delete your account");
    redirect("/account");
  }

  const cookieStore = await cookies();
  const result = await deleteCurrentUser(cookieStore.toString(), email);
  if (!result.ok) {
    await setFlash("error", "Unable to delete account");
    redirect("/account");
  }

  await logout(cookieStore.toString());
  await clearAuthCookies();
  await setFlash("success", "Account deleted");
  redirect("/");
}

export async function createClientAction(formData: FormData) {
	const { name, parsedRedirectURI, allowedScopes, error } = parseClientForm(formData);
	if (error || !parsedRedirectURI) {
		await setFlash("error", error ?? "Invalid client data");
		redirect("/account");
	}

	const cookieStore = await cookies();
	const result = await createCurrentUserClient(cookieStore.toString(), {
		name,
		redirectURI: parsedRedirectURI.toString(),
		allowedScopes,
	});
	if (!result.ok) {
		await setFlash("error", "Unable to create client");
		redirect("/account");
	}

	await setFlash("success", "Client created");
	redirect("/account");
}

export async function revokeAuthorizationAction(formData: FormData) {
	const clientId = String(formData.get("client_id") ?? "").trim();
	if (!clientId) {
		await setFlash("error", "Client ID is required");
		redirect("/account");
	}

	const cookieStore = await cookies();
	const result = await revokeCurrentUserAuthorization(cookieStore.toString(), clientId);
	if (!result.ok) {
		await setFlash("error", "Unable to disconnect app");
		redirect("/account");
	}

	await setFlash("success", "App disconnected");
	redirect("/account");
}

export async function updateClientAction(formData: FormData) {
	const clientId = String(formData.get("client_id") ?? "").trim();
	if (!clientId) {
		await setFlash("error", "Client ID is required");
		redirect("/account");
	}

	const { name, parsedRedirectURI, allowedScopes, error } = parseClientForm(formData);
	if (error || !parsedRedirectURI) {
		await setFlash("error", error ?? "Invalid client data");
		redirect("/account");
	}

	const cookieStore = await cookies();
	const result = await updateCurrentUserClient(cookieStore.toString(), clientId, {
		name,
		redirectURI: parsedRedirectURI.toString(),
		allowedScopes,
	});
	if (!result.ok) {
		await setFlash("error", "Unable to update client");
		redirect("/account");
	}

	await setFlash("success", "Client updated");
	redirect("/account");
}

type ParsedClientForm = {
	name: string;
	parsedRedirectURI?: URL;
	allowedScopes: string[];
	error: string | null;
};

function parseClientForm(formData: FormData): ParsedClientForm {

	const name = String(formData.get("name") ?? "").trim();
	const redirectURI = String(formData.get("redirect_uri") ?? "").trim();
	const allowedScopesInput = String(formData.get("allowed_scopes") ?? "").trim();

	if (!name || !redirectURI || !allowedScopesInput) {
		return {
			name: "",
			parsedRedirectURI: undefined,
			allowedScopes: [] as string[],
			error: "Name, redirect URI, and scopes are required",
		};
	}

	let parsedRedirectURI: URL;
	try {
		parsedRedirectURI = new URL(redirectURI);
	} catch {
		return {
			name: "",
			parsedRedirectURI: undefined,
			allowedScopes: [] as string[],
			error: "Redirect URI must be a valid URL",
		};
	}

	if (parsedRedirectURI.protocol !== "http:" && parsedRedirectURI.protocol !== "https:") {
		return {
			name: "",
			parsedRedirectURI: undefined,
			allowedScopes: [] as string[],
			error: "Redirect URI must use http or https",
		};
	}

	const allowedScopes = Array.from(
		new Set(
			allowedScopesInput
				.split(/\s+/)
				.map((scope) => scope.trim())
				.filter(Boolean),
		),
	);

	if (allowedScopes.length === 0) {
		return {
			name: "",
			parsedRedirectURI: undefined,
			allowedScopes: [] as string[],
			error: "At least one allowed scope is required",
		};
	}

	if (!allowedScopes.includes("openid")) {
		allowedScopes.unshift("openid");
	}

	return {
		name,
		parsedRedirectURI,
		allowedScopes,
		error: null,
	};
}

export async function revokeClientAction(formData: FormData) {
	const clientId = String(formData.get("client_id") ?? "").trim();
	if (!clientId) {
		await setFlash("error", "Client ID is required");
		redirect("/account");
	}

	const cookieStore = await cookies();
	const result = await revokeCurrentUserClient(cookieStore.toString(), clientId);
	if (!result.ok) {
		await setFlash("error", "Unable to revoke client");
		redirect("/account");
	}

	await setFlash("success", "Client revoked");
	redirect("/account");
}

export async function deleteClientAction(formData: FormData) {
	const clientId = String(formData.get("client_id") ?? "").trim();
	if (!clientId) {
		await setFlash("error", "Client ID is required");
		redirect("/account");
	}

	const cookieStore = await cookies();
	const result = await deleteCurrentUserClient(cookieStore.toString(), clientId);
	if (!result.ok) {
		await setFlash("error", "Unable to delete client");
		redirect("/account");
	}

	await setFlash("success", "Client deleted");
	redirect("/account");
}

export async function listCurrentUserClientsAction(cursor?: string, limit = 20) {
	const cookieStore = await cookies();
	const response = await listCurrentUserClients(cookieStore.toString(), {
		cursor,
		limit,
	});
	if (!response) {
		return { clients: [], nextCursor: undefined, error: "Unable to load clients" };
	}

	return { ...response, error: undefined };
}

export async function createRoleAction(formData: FormData) {
  const name = String(formData.get("name") ?? "").trim();
  if (!name) {
    await setFlash("error", "Role name is required");
    redirect("/admin");
  }

  const cookieStore = await cookies();
  const result = await createRole(cookieStore.toString(), name);
  if (!result.ok) {
    await setFlash("error", "Unable to create role");
    redirect("/admin");
  }

  await setFlash("success", "Role created");
  redirect("/admin");
}

export async function deleteRoleAction(formData: FormData) {
  const name = String(formData.get("name") ?? "").trim();
  if (!name) {
    await setFlash("error", "Role name is required");
    redirect("/admin");
  }

  const cookieStore = await cookies();
  const result = await deleteRole(cookieStore.toString(), name);
  if (!result.ok) {
    await setFlash("error", "Unable to delete role");
    redirect("/admin");
  }

  await setFlash("success", "Role deleted");
  redirect("/admin");
}

export async function addUserRoleAction(formData: FormData) {
  const userId = String(formData.get("user_id") ?? "").trim();
  const role = String(formData.get("role") ?? "").trim();
  if (!userId || !role) {
    await setFlash("error", "User and role are required");
    redirect("/admin");
  }

  const cookieStore = await cookies();
  const result = await addUserRole(cookieStore.toString(), userId, role);
  if (!result.ok) {
    await setFlash("error", "Unable to assign role");
    redirect("/admin");
  }

  await setFlash("success", "Role assigned");
  redirect("/admin");
}

export async function removeUserRoleAction(formData: FormData) {
  const userId = String(formData.get("user_id") ?? "").trim();
  const role = String(formData.get("role") ?? "").trim();
  if (!userId || !role) {
    await setFlash("error", "User and role are required");
    redirect("/admin");
  }

  const cookieStore = await cookies();
  const result = await removeUserRole(cookieStore.toString(), userId, role);
  if (!result.ok) {
    await setFlash("error", "Unable to remove role");
    redirect("/admin");
  }

  await setFlash("success", "Role removed");
  redirect("/admin");
}

export async function listAdminUserEventsAction(
  userId: string,
  cursor?: string,
  limit = 20,
) {
  if (!userId) {
    return { events: [], nextCursor: undefined, error: "User ID is required" };
  }

  const cookieStore = await cookies();
  const response = await listAdminUserEvents(cookieStore.toString(), userId, {
    cursor,
    limit,
  });
  if (!response) {
    return { events: [], nextCursor: undefined, error: "Unable to load events" };
  }

	return { ...mapAdminUserEvents(response), error: undefined };
}

export async function listAdminUserRefreshTokensAction(
	userId: string,
	cursor?: string,
	limit = 20,
) {
	if (!userId) {
		return { sessions: [], nextCursor: undefined, error: "User ID is required" };
	}

	const cookieStore = await cookies();
	const response = await listAdminUserRefreshTokens(cookieStore.toString(), userId, {
		cursor,
		limit,
	});
	if (!response) {
		return { sessions: [], nextCursor: undefined, error: "Unable to load sessions" };
	}

	return { ...mapAdminUserSessions(response), error: undefined };
}

export async function authorizeContinueAction(formData: FormData) {
  const callbackUrl = String(formData.get("callback_url") ?? "").trim();
  const state = String(formData.get("state") ?? "").trim();

  if (!callbackUrl || !isAllowedCallbackUrl(callbackUrl)) {
    redirect("/authorize");
  }

  const cookieStore = await cookies();
  const token = cookieStore.get("access_token")?.value;
  if (!token) {
    redirect(`/login?callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}`);
  }

  redirect(buildAuthorizeSuccessUrl(callbackUrl, state));
}
