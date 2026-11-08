// Session cookies shared by proxy.ts, Server Actions and Server Components.
// Tokens never reach client JavaScript: every cookie is httpOnly, and the browser only
// talks to this Next.js server (BFF), never to the Jazyl API directly.

const secure = process.env.COOKIE_SECURE
  ? process.env.COOKIE_SECURE === "true"
  : process.env.NODE_ENV === "production";

// "__Host-" pins the cookie to this exact host, HTTPS and path "/" (no subdomain can set it).
const prefix = secure ? "__Host-" : "";

export const COOKIE = {
  access: `${prefix}jz_at`,
  refresh: `${prefix}jz_rt`,
  mfa: `${prefix}jz_mfa`,
} as const;

export type CookieOptions = {
  httpOnly: true;
  secure: boolean;
  sameSite: "lax";
  path: "/";
  maxAge: number;
};

export function cookieOptions(maxAgeSeconds: number): CookieOptions {
  return { httpOnly: true, secure, sameSite: "lax", path: "/", maxAge: maxAgeSeconds };
}

/**
 * Expires a session cookie. `__Host-` cookies are only replaced when the deleting Set-Cookie
 * carries the same Secure/Path attributes, so a bare `.delete(name)` would silently fail.
 */
export function expireCookie(jar: { set: (name: string, value: string, opts: CookieOptions) => unknown }, name: string) {
  jar.set(name, "", cookieOptions(0));
}

export const REFRESH_MAX_AGE = 30 * 24 * 60 * 60;

export type TokenPair = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  mfa: boolean;
  mfa_setup_required: boolean;
  must_change_password: boolean;
  user: { id: string; role: string; clinic_id: string | null };
};

export type AccessClaims = {
  sub: string;
  role: "superadmin" | "owner" | "admin";
  cid?: string;
  mfa?: boolean;
  pwc?: boolean;
  exp: number;
};

/** Reads the JWT payload WITHOUT verifying it. Only for routing hints; the API verifies. */
export function peekClaims(token: string | undefined): AccessClaims | null {
  if (!token) return null;
  const part = token.split(".")[1];
  if (!part) return null;
  try {
    const json = atob(part.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(json) as AccessClaims;
  } catch {
    return null;
  }
}

/** True when the access token is missing or expires within `skewSeconds`. */
export function accessExpired(token: string | undefined, skewSeconds = 30): boolean {
  const c = peekClaims(token);
  return !c || c.exp * 1000 - skewSeconds * 1000 <= Date.now();
}

/** Where a freshly authenticated session should land. */
export function landingFor(t: Pick<TokenPair, "must_change_password" | "mfa_setup_required">, next?: string | null) {
  if (t.must_change_password) return "/setup/password";
  if (t.mfa_setup_required) return "/setup/2fa";
  return safeNext(next) ?? "/";
}

/** Only same-site absolute paths are accepted as a post-login destination (no open redirect). */
export function safeNext(next: string | null | undefined): string | null {
  if (!next || !next.startsWith("/") || next.startsWith("//") || next.startsWith("/\\")) return null;
  if (/^\/(login|setup|auth)(\/|\?|$)/.test(next)) return null;
  return next;
}
