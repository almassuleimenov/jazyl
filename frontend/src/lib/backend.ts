// Low-level calls to the Jazyl Go API. Usable from proxy.ts (no next/headers here).

export function apiBase(): string {
  const base = process.env.JAZYL_API_URL;
  if (!base) throw new Error("JAZYL_API_URL is not set");
  return base.replace(/\/+$/, "");
}

export type ApiErrorBody = { error?: { code?: string; message?: string; request_id?: string } };

export type RawResult = { status: number; data: unknown; retryAfter: string | null };

export async function rawCall(
  path: string,
  init: { method?: string; body?: unknown; token?: string; clientIp?: string | null; userAgent?: string | null } = {},
): Promise<RawResult> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (init.body !== undefined) headers["Content-Type"] = "application/json";
  if (init.token) headers.Authorization = `Bearer ${init.token}`;
  // The API rate-limits login per client IP; forward the real one (API needs TRUSTED_PROXY=true).
  if (init.clientIp) headers["X-Forwarded-For"] = init.clientIp;
  if (init.userAgent) headers["User-Agent"] = init.userAgent.slice(0, 300);

  let res: Response;
  try {
    res = await fetch(`${apiBase()}/api/v1${path}`, {
      method: init.method ?? "GET",
      headers,
      body: init.body === undefined ? undefined : JSON.stringify(init.body),
      cache: "no-store",
      signal: AbortSignal.timeout(15_000),
    });
  } catch (e) {
    // API down or timed out: report it like a 503 so forms show a message instead of crashing.
    console.error(`Jazyl API unreachable (${apiBase()}):`, (e as Error).cause ?? e);
    return { status: 503, data: { error: { code: "unavailable", message: "API unreachable" } }, retryAfter: null };
  }
  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: { code: "bad_response", message: text.slice(0, 200) } };
    }
  }
  return { status: res.status, data, retryAfter: res.headers.get("Retry-After") };
}

export function errorCode(data: unknown): string | undefined {
  return (data as ApiErrorBody | null)?.error?.code;
}

export function firstForwardedIp(xff: string | null): string | null {
  if (!xff) return null;
  return xff.split(",")[0]?.trim() || null;
}
