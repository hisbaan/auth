"use server";

import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";

import { buildAuthorizeSuccessUrl, isAllowedCallbackUrl } from "@/lib/callback";
import { COOKIE_DOMAIN, COOKIE_SECURE } from "@/lib/config";
import {
  addUserRole,
  createRole,
  deleteCurrentUser,
  deleteRole,
  listAdminUserEvents,
  loginWithPassword,
  logout,
  registerUser,
  removeUserRole,
  resetPassword,
  sendForgotPassword,
  updateCurrentPassword,
  updateCurrentUser,
  verifyEmail,
} from "@/lib/sdk";
import { mapAdminUserEvents } from "@/lib/event-utils";
import { setFlash } from "@/lib/flash";
import { sanitizeRelativePath } from "@/lib/utils";

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

  if (!email || !password) {
    await setFlash("error", "Invalid credentials");
    redirect(`/login?next=${encodeURIComponent(next)}&callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}`);
  }

  const requestHeaders = await headers();
  const userAgent = requestHeaders.get("user-agent") ?? undefined;
  const result = await loginWithPassword(email, password, userAgent);
  if (!result.ok || !result.data) {
    await setFlash("error", "Invalid credentials");
    redirect(`/login?next=${encodeURIComponent(next)}&callback_url=${encodeURIComponent(callbackUrl)}&state=${encodeURIComponent(state)}`);
  }

  const cookieStore = await cookies();
  cookieStore.set("access_token", result.data.access_token, cookieOptions(result.data.expires_in));
  cookieStore.set("refresh_token", result.data.refresh_token, cookieOptions(60 * 60 * 24 * 7));

  let destination = sanitizeRelativePath(next, "/account");
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

export async function verifyEmailAction(formData: FormData) {
  const token = String(formData.get("token") ?? "").trim();
  if (!token) {
    await setFlash("error", "Token is required");
    redirect("/verify-email");
  }

  const result = await verifyEmail(token);
  if (!result.ok) {
    await setFlash("error", "Unable to verify email");
    redirect(`/verify-email?token=${encodeURIComponent(token)}`);
  }

  await setFlash("success", "Email verified. You can now sign in.");
  redirect("/verify-email");
}

export async function logoutAction() {
  const cookieStore = await cookies();
  await logout(cookieStore.toString());
  await clearAuthCookies();
  redirect("/");
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
  if (!result.ok) {
    await setFlash("error", "Unable to update profile");
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

export async function deleteAccountAction() {
  const cookieStore = await cookies();
  const result = await deleteCurrentUser(cookieStore.toString());
  if (!result.ok) {
    await setFlash("error", "Unable to delete account");
    redirect("/account");
  }

  await logout(cookieStore.toString());
  await clearAuthCookies();
  await setFlash("success", "Account deleted");
  redirect("/");
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
