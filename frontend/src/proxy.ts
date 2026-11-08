import { NextResponse, type NextRequest } from "next/server";

import { errorCode, firstForwardedIp, rawCall } from "@/lib/backend";
import { COOKIE, REFRESH_MAX_AGE, accessExpired, cookieOptions, type TokenPair, expireCookie } from "@/lib/session";

// Proxy keeps the short-lived access token fresh before any page or Server Action runs,
// and sends anonymous visitors to /login. Authorization itself is enforced by the Jazyl API
// on every call; this layer only manages the session cookies.

type Refreshed = { ok: true; tokens: TokenPair } | { ok: false; race: boolean; down?: boolean };

// Parallel requests (prefetches) often carry the same expiring refresh token. Rotation is
// single-use, so share one in-flight refresh per token inside this server process.
const inflight = new Map<string, { at: number; promise: Promise<Refreshed> }>();
const SHARE_MS = 20_000;

function refreshOnce(rt: string, clientIp: string | null, userAgent: string | null): Promise<Refreshed> {
  const now = Date.now();
  for (const [k, v] of inflight) if (now - v.at > SHARE_MS) inflight.delete(k);
  const hit = inflight.get(rt);
  if (hit) return hit.promise;
  const promise = rawCall("/auth/refresh", { method: "POST", body: { refresh_token: rt }, clientIp, userAgent })
    .then((r): Refreshed => {
      if (r.status === 200) return { ok: true, tokens: r.data as TokenPair };
      if (r.status >= 500) return { ok: false, race: false, down: true }; // API down: keep the session
      return { ok: false, race: errorCode(r.data) === "refresh_race" };
    })
    .catch((): Refreshed => ({ ok: false, race: false, down: true })); // API unreachable: keep the session
  inflight.set(rt, { at: now, promise });
  return promise;
}

function toLogin(req: NextRequest, clear: boolean) {
  const url = req.nextUrl.clone();
  const next = req.nextUrl.pathname + req.nextUrl.search;
  url.pathname = "/login";
  url.search = next !== "/" ? `?next=${encodeURIComponent(next)}` : "";
  const res = NextResponse.redirect(url);
  if (clear) {
    expireCookie(res.cookies, COOKIE.access);
    expireCookie(res.cookies, COOKIE.refresh);
  }
  return res;
}

export async function proxy(req: NextRequest) {
  const isLogin = req.nextUrl.pathname === "/login";
  const at = req.cookies.get(COOKIE.access)?.value;
  const rt = req.cookies.get(COOKIE.refresh)?.value;

  if (!accessExpired(at)) {
    return isLogin ? NextResponse.redirect(new URL("/", req.url)) : NextResponse.next();
  }
  if (!rt) {
    return isLogin ? NextResponse.next() : toLogin(req, Boolean(at));
  }

  const r = await refreshOnce(rt, firstForwardedIp(req.headers.get("x-forwarded-for")), req.headers.get("user-agent"));
  if (!r.ok) {
    if (r.down) {
      // Let the page render; its data calls fail into the error boundary instead of a redirect loop.
      inflight.delete(rt);
      return NextResponse.next();
    }
    if (r.race && req.method === "GET" && !isLogin) {
      // Another request of this browser just rotated the token; its cookies arrive momentarily.
      return NextResponse.redirect(req.nextUrl);
    }
    if (isLogin) {
      const res = NextResponse.next();
      expireCookie(res.cookies, COOKIE.access);
      expireCookie(res.cookies, COOKIE.refresh);
      return res;
    }
    return toLogin(req, true);
  }

  // Hand the fresh access token to this very render, and store both tokens in the browser.
  const fresh = new Headers(req.headers);
  req.cookies.set(COOKIE.access, r.tokens.access_token);
  req.cookies.set(COOKIE.refresh, r.tokens.refresh_token);
  fresh.set("cookie", req.cookies.toString());
  const res = isLogin ? NextResponse.redirect(new URL("/", req.url)) : NextResponse.next({ request: { headers: fresh } });
  res.cookies.set(COOKIE.access, r.tokens.access_token, cookieOptions(r.tokens.expires_in));
  res.cookies.set(COOKIE.refresh, r.tokens.refresh_token, cookieOptions(REFRESH_MAX_AGE));
  return res;
}

export const config = {
  // Everything except static assets and the setup pages' own static files.
  matcher: ["/((?!_next/static|_next/image|favicon.ico|icon.svg|robots.txt).*)"],
};
