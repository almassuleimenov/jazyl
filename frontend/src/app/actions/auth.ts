"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import QRCode from "qrcode";

import { ApiError, api, clearSession, humanError, publicApi, storeSession, type ActionState } from "@/lib/api";
import { COOKIE, cookieOptions, landingFor, safeNext, type TokenPair, expireCookie } from "@/lib/session";

export type LoginState = { step: "password" | "mfa"; error?: string; email?: string; next?: string } | undefined;

type LoginResponse = TokenPair | { mfa_required: true; mfa_token: string; expires_in: number };

const str = (fd: FormData, k: string) => String(fd.get(k) ?? "").trim();

export async function loginAction(_: LoginState, fd: FormData): Promise<LoginState> {
  const email = str(fd, "email");
  const password = String(fd.get("password") ?? "");
  const next = safeNext(str(fd, "next")) ?? undefined;
  if (!email || !password) return { step: "password", email, next, error: "Введите почту и пароль." };

  let res: LoginResponse;
  try {
    res = await publicApi<LoginResponse>("/auth/login", { email, password });
  } catch (e) {
    if (e instanceof ApiError) return { step: "password", email, next, error: humanError(e) };
    throw e;
  }
  if ("mfa_required" in res) {
    (await cookies()).set(COOKIE.mfa, res.mfa_token, cookieOptions(res.expires_in));
    return { step: "mfa", email, next };
  }
  await storeSession(res);
  redirect(landingFor(res, next));
}

export async function mfaAction(prev: LoginState, fd: FormData): Promise<LoginState> {
  const code = str(fd, "code").replace(/\s/g, "");
  const next = safeNext(str(fd, "next")) ?? undefined;
  const mfaToken = (await cookies()).get(COOKIE.mfa)?.value;
  if (!mfaToken) return { step: "password", email: prev?.email, next, error: "Время на ввод кода истекло. Войдите ещё раз." };
  if (!/^\d{6}$/.test(code)) return { step: "mfa", email: prev?.email, next, error: "Введите 6 цифр из приложения." };

  let res: TokenPair;
  try {
    res = await publicApi<TokenPair>("/auth/mfa", { mfa_token: mfaToken, code });
  } catch (e) {
    if (e instanceof ApiError) {
      const restart = e.code === "invalid_token" || e.code === "account_locked";
      if (restart) expireCookie(await cookies(), COOKIE.mfa);
      return { step: restart ? "password" : "mfa", email: prev?.email, next, error: humanError(e) };
    }
    throw e;
  }
  await storeSession(res);
  redirect(landingFor(res, next));
}

export async function logoutAction() {
  const rt = (await cookies()).get(COOKIE.refresh)?.value;
  if (rt) {
    try {
      await publicApi("/auth/logout", { refresh_token: rt });
    } catch {
      // The cookie is dropped anyway; an unrevoked token expires on its own.
    }
  }
  await clearSession();
  redirect("/login");
}

export async function logoutAllAction() {
  await api("/auth/logout-all", { method: "POST" });
  await clearSession();
  redirect("/login?signed_out=all");
}

export async function changePasswordAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const current = String(fd.get("current_password") ?? "");
  const next = String(fd.get("new_password") ?? "");
  const confirm = String(fd.get("confirm_password") ?? "");
  if (next.length < 10) return { ok: false, error: "Новый пароль должен быть не короче 10 символов." };
  if (next !== confirm) return { ok: false, error: "Пароли не совпадают." };

  let res: TokenPair;
  try {
    res = await api<TokenPair>("/auth/password", { method: "POST", body: { current_password: current, new_password: next } });
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
  await storeSession(res);
  if (fd.get("forced") === "1") redirect(landingFor(res));
  return { ok: true, message: "Пароль изменён. На других устройствах нужно войти заново." };
}

export type TotpSetup = { secret: string; qr: string } | { error: string };

export async function startTotpAction(): Promise<TotpSetup> {
  try {
    const r = await api<{ secret: string; otpauth_url: string }>("/auth/totp/setup", { method: "POST" });
    const qr = await QRCode.toString(r.otpauth_url, {
      type: "svg",
      margin: 0,
      errorCorrectionLevel: "M",
      color: { dark: "#1f2328", light: "#ffffff" },
    });
    return { secret: r.secret, qr };
  } catch (e) {
    if (e instanceof ApiError) return { error: humanError(e) };
    throw e;
  }
}

export async function enableTotpAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const code = str(fd, "code").replace(/\s/g, "");
  if (!/^\d{6}$/.test(code)) return { ok: false, error: "Введите 6 цифр из приложения." };
  let res: TokenPair;
  try {
    res = await api<TokenPair>("/auth/totp/enable", { method: "POST", body: { code } });
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
  await storeSession(res);
  if (fd.get("forced") === "1") redirect(landingFor(res));
  return { ok: true, message: "Двухфакторная защита включена." };
}

export async function disableTotpAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const code = str(fd, "code").replace(/\s/g, "");
  const password = String(fd.get("password") ?? "");
  let res: TokenPair;
  try {
    res = await api<TokenPair>("/auth/totp/disable", { method: "POST", body: { password, code } });
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
  await storeSession(res);
  return { ok: true, message: "Двухфакторная защита выключена." };
}
