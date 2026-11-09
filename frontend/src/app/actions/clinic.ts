"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { ApiError, api, apiWithStatus, attempt, humanError, type ActionState } from "@/lib/api";
import { localInputToISO } from "@/lib/format";
import type { Clinic, ClinicSettings } from "@/lib/types";

// Every action re-checks access on the API: the clinic id comes from the form, and the API
// answers 404 for clinics outside the caller's tenant. Nothing here is trusted client-side.

const s = (fd: FormData, k: string) => String(fd.get(k) ?? "").trim();
const n = (fd: FormData, k: string) => Number(s(fd, k));
const b = (fd: FormData, k: string) => fd.get(k) === "on";
const id = (fd: FormData, k = "clinic_id") => encodeURIComponent(s(fd, k));
const phones = (raw: string) =>
  raw
    .split(/[\n,;]+/)
    .map((p) => p.trim())
    .filter(Boolean);

// ---------- platform ----------

export async function createClinicAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const body = {
    name: s(fd, "name"),
    slug: s(fd, "slug").toLowerCase(),
    city: s(fd, "city"),
    address: s(fd, "address"),
    plan: s(fd, "plan"),
    status: s(fd, "status"),
    owner_phone: s(fd, "owner_phone"),
    admin_phones: phones(s(fd, "admin_phones")),
    gis_review_url: s(fd, "gis_review_url"),
    avg_check_kzt: n(fd, "avg_check_kzt") || 15000,
    wa_provider: s(fd, "wa_provider") || "openwa",
  };
  let c: Clinic;
  try {
    c = await api<Clinic>("/admin/clinics", { method: "POST", body });
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: e.status === 409 ? "Клиника с таким коротким именем уже есть." : humanError(e) };
    throw e;
  }
  revalidatePath("/admin");
  redirect(`/c/${c.id}/team?created=1`);
}

export async function connectWhatsAppAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() => api(`/admin/clinics/${id(fd)}/whatsapp/connect`, { method: "POST" }));
  revalidatePath(`/c/${s(fd, "clinic_id")}/settings`);
  return r?.ok ? { ok: true, message: "Сессия запущена. Отсканируйте QR-код телефоном клиники." } : r;
}

export type QrState = { qr?: string; status?: string; code?: string; error?: string } | undefined;

export async function whatsappQrAction(clinicId: string): Promise<QrState> {
  try {
    const r = await api<{ qr_code: string; status: string }>(`/admin/clinics/${encodeURIComponent(clinicId)}/whatsapp/qr`);
    // Only accept an image data URL from the gateway (rendered as <img>, never as HTML).
    const qr = /^data:image\/(png|jpeg|svg\+xml);base64,[A-Za-z0-9+/=]+$/.test(r.qr_code) ? r.qr_code : undefined;
    return { qr, status: r.status };
  } catch (e) {
    if (e instanceof ApiError) return { error: humanError(e) };
    throw e;
  }
}

export async function pairingCodeAction(_: QrState, fd: FormData): Promise<QrState> {
  try {
    const r = await api<{ pairing_code: string }>(`/admin/clinics/${id(fd)}/whatsapp/pairing-code`, {
      method: "POST",
      body: { phone: s(fd, "phone") },
    });
    return { code: r.pairing_code };
  } catch (e) {
    if (e instanceof ApiError) return { error: humanError(e) };
    throw e;
  }
}

// ---------- clinic profile & settings ----------

export async function updateClinicAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const clinicId = s(fd, "clinic_id");
  const current = await api<Clinic>(`/clinics/${id(fd)}/`);
  const settings: ClinicSettings = {
    ...current.settings,
    reminder_evening_hour: n(fd, "reminder_evening_hour"),
    two_hour_reminder: b(fd, "two_hour_reminder"),
    quiet_start_hour: n(fd, "quiet_start_hour"),
    quiet_end_hour: n(fd, "quiet_end_hour"),
    chairs: n(fd, "chairs"),
    slot_minutes: n(fd, "slot_minutes"),
    auto_book_from_site: b(fd, "auto_book_from_site"),
    auto_complete_confirmed: b(fd, "auto_complete_confirmed"),
    review_requests: b(fd, "review_requests"),
    weekly_report: b(fd, "weekly_report"),
    monthly_report: b(fd, "monthly_report"),
  };
  const body: Record<string, unknown> = {
    name: s(fd, "name"),
    city: s(fd, "city"),
    address: s(fd, "address"),
    owner_phone: s(fd, "owner_phone"),
    admin_phones: phones(s(fd, "admin_phones")),
    gis_review_url: s(fd, "gis_review_url"),
    avg_check_kzt: n(fd, "avg_check_kzt"),
    language: s(fd, "language") || current.language,
    settings,
  };
  // Plan, status and provider are Jazyl's; the form only sends them for the platform role.
  if (fd.has("plan")) {
    body.plan = s(fd, "plan");
    body.status = s(fd, "status");
    body.wa_provider = s(fd, "wa_provider");
  }
  const r = await attempt(() => api(`/clinics/${id(fd)}/`, { method: "PATCH", body }), "Настройки сохранены.");
  revalidatePath(`/c/${clinicId}`, "layout");
  return r;
}

// ---------- team & access ----------

export async function createUserAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(
    () =>
      api(`/clinics/${id(fd)}/users`, {
        method: "POST",
        body: { email: s(fd, "email"), name: s(fd, "name"), role: s(fd, "role"), password: String(fd.get("password") ?? "") },
      }),
    "Доступ создан. Передайте почту и временный пароль лично: при первом входе система попросит сменить пароль.",
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}/team`);
  return r;
}

export async function setUserActiveAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() =>
    api(`/clinics/${id(fd)}/users/${id(fd, "user_id")}`, { method: "PATCH", body: { is_active: s(fd, "active") === "1" } }),
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}/team`);
  return r;
}

export async function resetUserPasswordAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(
    () => api(`/clinics/${id(fd)}/users/${id(fd, "user_id")}/password`, { method: "POST", body: { password: String(fd.get("password") ?? "") } }),
    "Новый временный пароль выдан. Все сессии сотрудника закрыты, при входе он сменит пароль.",
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}/team`);
  return r;
}

export async function resetUserTotpAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() => api(`/clinics/${id(fd)}/users/${id(fd, "user_id")}/reset-2fa`, { method: "POST" }));
  revalidatePath(`/c/${s(fd, "clinic_id")}/team`);
  return r;
}

// ---------- site keys ----------

export type SiteKeyState = (ActionState & { key?: string }) | undefined;

export async function createSiteKeyAction(_: SiteKeyState, fd: FormData): Promise<SiteKeyState> {
  const origins = s(fd, "allowed_origins")
    .split(/[\s,]+/)
    .map((o) => o.trim().replace(/\/+$/, ""))
    .filter(Boolean);
  try {
    const r = await api<{ key: string }>(`/clinics/${id(fd)}/site-keys`, {
      method: "POST",
      body: { name: s(fd, "name"), allowed_origins: origins },
    });
    revalidatePath(`/c/${s(fd, "clinic_id")}/site`);
    return { ok: true, key: r.key };
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
}

export async function revokeSiteKeyAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() => api(`/clinics/${id(fd)}/site-keys/${id(fd, "key_id")}`, { method: "DELETE" }));
  revalidatePath(`/c/${s(fd, "clinic_id")}/site`);
  return r;
}

// ---------- appointments, leads, messages ----------

export async function setAppointmentStatusAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() =>
    api(`/clinics/${id(fd)}/appointments/${id(fd, "appointment_id")}/status`, { method: "POST", body: { status: s(fd, "status") } }),
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}`, "layout");
  return r;
}

export async function createAppointmentAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const date = s(fd, "date");
  const timeOfDay = s(fd, "time");
  if (!date || !timeOfDay) return { ok: false, error: "Укажите дату и время." };
  let status: number;
  try {
    ({ status } = await apiWithStatus(`/clinics/${id(fd)}/appointments`, {
      method: "POST",
      body: {
        patient_phone: s(fd, "patient_phone"),
        patient_name: s(fd, "patient_name"),
        patient_language: s(fd, "patient_language") || "ru",
        service_name: s(fd, "service_name"),
        starts_at: localInputToISO(`${date}T${timeOfDay}`),
        duration_min: n(fd, "duration_min") || 30,
        notes: s(fd, "notes"),
      },
    }));
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
  // The API never duplicates a visit: same patient + same start returns the existing one with 200.
  if (status !== 201) {
    const [y, m, d] = date.split("-");
    return {
      ok: false,
      error: `Этот пациент уже записан на ${d}.${m}.${y} в ${timeOfDay}. Повторная запись не создаётся, чтобы напоминание не пришло дважды.`,
    };
  }
  revalidatePath(`/c/${s(fd, "clinic_id")}`, "layout");
  // Open the day of the new visit: the list defaults to "tomorrow" and would otherwise hide it.
  redirect(`/c/${id(fd)}/appointments?day=${encodeURIComponent(date)}`);
}

export async function importScheduleAction(_: ActionState, fd: FormData): Promise<ActionState> {
  try {
    const r = await api<{ summary: string }>(`/clinics/${id(fd)}/schedule/import`, { method: "POST", body: { text: s(fd, "text") } });
    revalidatePath(`/c/${s(fd, "clinic_id")}`, "layout");
    return { ok: true, message: r.summary };
  } catch (e) {
    if (e instanceof ApiError) return { ok: false, error: humanError(e) };
    throw e;
  }
}

export async function updateLeadAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(() => api(`/clinics/${id(fd)}/leads/${id(fd, "lead_id")}`, { method: "PATCH", body: { status: s(fd, "status") } }));
  revalidatePath(`/c/${s(fd, "clinic_id")}`, "layout");
  return r;
}

export async function convertLeadAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const date = s(fd, "date");
  const timeOfDay = s(fd, "time");
  if (!date || !timeOfDay) return { ok: false, error: "Укажите дату и время приёма." };
  const r = await attempt(
    () =>
      api(`/clinics/${id(fd)}/leads/${id(fd, "lead_id")}/convert`, {
        method: "POST",
        body: { starts_at: localInputToISO(`${date}T${timeOfDay}`) },
      }),
    "Пациент записан, напоминания запланированы.",
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}`, "layout");
  return r;
}

export async function sendMessageAction(_: ActionState, fd: FormData): Promise<ActionState> {
  const r = await attempt(
    () => api(`/clinics/${id(fd)}/messages`, { method: "POST", body: { phone: s(fd, "phone"), text: s(fd, "text") } }),
    "Сообщение поставлено в очередь отправки.",
  );
  revalidatePath(`/c/${s(fd, "clinic_id")}/chats`);
  return r;
}
