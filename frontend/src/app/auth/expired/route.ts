import { NextResponse, type NextRequest } from "next/server";

import { COOKIE, expireCookie } from "@/lib/session";

// The API rejected a token that still looks valid (revoked by "log out everywhere", a password
// reset by Jazyl, or deactivation). Drop the cookies so /login does not bounce back.
export function GET(req: NextRequest) {
  const res = NextResponse.redirect(new URL("/login?expired=1", req.url));
  expireCookie(res.cookies, COOKIE.access);
  expireCookie(res.cookies, COOKIE.refresh);
  expireCookie(res.cookies, COOKIE.mfa);
  return res;
}
