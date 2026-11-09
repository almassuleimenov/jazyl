import "server-only";

import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";

import { errorCode, firstForwardedIp, rawCall, type ApiErrorBody, type RawResult } from "./backend";
import { COOKIE, REFRESH_MAX_AGE, cookieOptions, type TokenPair, expireCookie } from "./session";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public retryAfter: string | null = null,
  ) {
    super(message);
  }
}

async function clientMeta() {
  const h = await headers();
  return {
    clientIp: firstForwardedIp(h.get("x-forwarded-for")) ?? h.get("x-real-ip"),
    userAgent: h.get("user-agent"),
  };
}

function toError(r: RawResult): ApiError {
  const body = r.data as ApiErrorBody | null;
  return new ApiError(r.status, body?.error?.code ?? "unknown", body?.error?.message ?? `HTTP ${r.status}`, r.retryAfter);
}

/**
 * Authenticated call to the Jazyl API with the session cookie.
 * Session problems end in a redirect (login, forced password change, 2FA setup);
 * other failures throw ApiError.
 */
export async function api<T>(path: string, init: { method?: string; body?: unknown } = {}): Promise<T> {
  return (await apiWithStatus<T>(path, init)).data;
}

/** Like api(), but also returns the HTTP status (e.g. 201 created vs 200 already existed). */
export async function apiWithStatus<T>(
  path: string,
  init: { method?: string; body?: unknown } = {},
): Promise<{ status: number; data: T }> {
  const token = (await cookies()).get(COOKIE.access)?.value;
  if (!token) redirect("/login");
  const r = await rawCall(path, { ...init, token, ...(await clientMeta()) });
  if (r.status >= 200 && r.status < 300) return { status: r.status, data: r.data as T };
  const code = errorCode(r.data);
  if (r.status === 401) redirect("/auth/expired");
  if (code === "password_change_required") redirect("/setup/password");
  if (code === "mfa_required") redirect("/setup/2fa");
  throw toError(r);
}

/** Unauthenticated call (login, 2FA step, logout). */
export async function publicApi<T>(path: string, body: unknown): Promise<T> {
  const r = await rawCall(path, { method: "POST", body, ...(await clientMeta()) });
  if (r.status >= 200 && r.status < 300) return r.data as T;
  throw toError(r);
}

/** Stores a token pair in httpOnly cookies. Only callable from Server Actions / Route Handlers. */
export async function storeSession(t: TokenPair) {
  const jar = await cookies();
  jar.set(COOKIE.access, t.access_token, cookieOptions(t.expires_in));
  jar.set(COOKIE.refresh, t.refresh_token, cookieOptions(REFRESH_MAX_AGE));
  expireCookie(jar, COOKIE.mfa);
}

export async function clearSession() {
  const jar = await cookies();
  expireCookie(jar, COOKIE.access);
  expireCookie(jar, COOKIE.refresh);
  expireCookie(jar, COOKIE.mfa);
}

// ---------- results for Server Actions ----------

export type ActionState = { ok?: boolean; error?: string; message?: string; fields?: Record<string, string> } | undefined;

const messages: Record<string, string> = {
  invalid_credentials: "Неверная почта или пароль.",
  invalid_code: "Код не подошёл. Введите свежий код из приложения; если не помогает, проверьте время на телефоне.",
  invalid_token: "Время на ввод кода истекло. Войдите ещё раз.",
  rate_limited: "Слишком много запросов. Подождите минуту и попробуйте снова.",
  conflict: "Такая запись уже существует.",
  forbidden: "Недостаточно прав для этого действия.",
  not_found: "Не найдено. Возможно, запись уже удалена.",
  opted_out: "Пациент отписался от сообщений, писать ему нельзя.",
  not_configured: "Сервер WhatsApp не настроен (OPENWA_WEBHOOK_SECRET).",
  unavailable: "Сервер Jazyl недоступен. Проверьте, что API запущен, и попробуйте ещё раз.",
  clinic_suspended: "Доступ к кабинету клиники приостановлен. Свяжитесь с командой Jazyl.",
};

// Known API validation messages translated for the interface.
const detail: [RegExp, string][] = [
  [/password must be at least/, "Пароль должен быть не короче 10 символов."],
  [/password must be at most/, "Пароль слишком длинный (до 72 символов)."],
  [/new password must differ/, "Новый пароль должен отличаться от текущего."],
  [/current password is wrong|password is wrong/, "Текущий пароль указан неверно."],
  [/already exists/, "Пользователь с такой почтой уже есть."],
  [/phone/, "Проверьте номер телефона: нужен казахстанский номер, например +7 701 123 45 67."],
  [/slug/, "Короткое имя: только латиница, цифры и дефис (например, aq-dent)."],
  [/allowed_origins/, "Домены сайта должны начинаться с https://"],
  [/gis_review_url/, "Ссылка на 2ГИС должна начинаться с https://"],
  [/name is required/, "Укажите название."],
  [/email/, "Проверьте адрес почты."],
  [/starts_at/, "Проверьте дату и время записи."],
  [/only Jazyl/, "Это может сделать только команда Jazyl."],
  [/cannot deactivate yourself/, "Нельзя отключить собственный доступ."],
  [/mandatory for Jazyl/, "Для аккаунта Jazyl двухфакторная защита обязательна."],
  [/start the setup first/, "Сначала начните подключение заново."],
  [/confirm must equal the clinic slug/, "Короткое имя введено неверно. Клиника не удалена."],
];

export function humanError(e: ApiError): string {
  if (e.code === "account_locked") {
    const mins = e.retryAfter ? Math.max(1, Math.ceil(Number(e.retryAfter) / 60)) : 15;
    return `Слишком много неудачных попыток. Вход заблокирован на ${mins} мин.`;
  }
  if (e.code === "invalid_credentials" && /current password|password is wrong/.test(e.message)) {
    return "Текущий пароль указан неверно.";
  }
  // Validation details only refine generic codes; a known code (e.g. a wrong login) wins.
  if (e.code === "invalid" || e.code === "conflict" || e.code === "forbidden") {
    for (const [re, text] of detail) if (re.test(e.message)) return text;
  }
  if (e.code === "whatsapp_error") {
    // The gateway's own reason (wrong key, rejected webhook URL, session not ready) is what the operator needs.
    return `WhatsApp-шлюз ответил ошибкой: ${e.message.slice(0, 300)}`;
  }
  if (messages[e.code]) return messages[e.code];
  if (e.status >= 500) return "Сервер Jazyl не ответил. Попробуйте ещё раз через минуту.";
  return "Не получилось сохранить. Проверьте поля и попробуйте снова.";
}

/** Runs fn and turns an expected ApiError into an action state; redirects and bugs propagate. */
export async function attempt(fn: () => Promise<unknown>, success?: string): Promise<ActionState> {
  try {
    await fn();
    return { ok: true, message: success };
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
}
