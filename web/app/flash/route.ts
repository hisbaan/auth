import { NextRequest, NextResponse } from "next/server";

import { setFlash } from "@/lib/flash";
import { sanitizeRelativePath } from "@/lib/utils";

const MAX_MESSAGE_LENGTH = 300;

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const type = searchParams.get("type");
  const message = searchParams.get("message");
  const redirectTo = searchParams.get("redirect") ?? "/";

  const destination = sanitizeRelativePath(redirectTo, "/");

  if ((type !== "success" && type !== "error") || !message) {
    return NextResponse.redirect(new URL(destination, request.url));
  }

  await setFlash(type, message.slice(0, MAX_MESSAGE_LENGTH));
  return NextResponse.redirect(new URL(destination, request.url));
}
