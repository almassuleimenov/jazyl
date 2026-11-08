import type { ApptStatus, ClinicStatus, LeadStatus, Plan, Role } from "./types";

export type Tone = "neutral" | "teal" | "amber" | "red" | "green";

export const apptStatus: Record<ApptStatus, { label: string; tone: Tone }> = {
  scheduled: { label: "Ждёт ответа", tone: "neutral" },
  confirmed: { label: "Подтвердил", tone: "green" },
  reschedule_requested: { label: "Просит перенос", tone: "amber" },
  cancelled: { label: "Отменена", tone: "red" },
  completed: { label: "Был на приёме", tone: "teal" },
  no_show: { label: "Не пришёл", tone: "red" },
};

export const leadStatus: Record<LeadStatus, { label: string; tone: Tone }> = {
  new: { label: "Новая", tone: "amber" },
  contacted: { label: "Связались", tone: "neutral" },
  booked: { label: "Записан", tone: "green" },
  rejected: { label: "Отказ", tone: "red" },
};

export const plans: Record<Plan, { label: string; price: number }> = {
  no_show: { label: "Без неявок", price: 15000 },
  full_chair: { label: "Полное кресло", price: 29000 },
  growth: { label: "Рост", price: 49000 },
};

export const clinicStatus: Record<ClinicStatus, { label: string; tone: Tone }> = {
  pilot: { label: "Пилот", tone: "amber" },
  active: { label: "Платит", tone: "green" },
  paused: { label: "Пауза", tone: "neutral" },
  churned: { label: "Ушла", tone: "red" },
};

export const roles: Record<Role, string> = {
  superadmin: "Команда Jazyl",
  owner: "Владелец",
  admin: "Администратор",
};

export function waStatus(s: string): { label: string; tone: Tone } {
  switch (s) {
    case "ready":
    case "connected":
    case "authenticated":
      return { label: "Подключён", tone: "green" };
    case "qr":
    case "qr_ready":
    case "scan_qr":
    case "starting":
    case "initializing":
      return { label: "Ждёт QR", tone: "amber" };
    case "disconnected":
    case "failed":
    case "logged_out":
      return { label: "Отключён", tone: "red" };
    case "":
    case "none":
      return { label: "Не подключён", tone: "neutral" };
    default:
      return { label: s, tone: "neutral" };
  }
}

export const leadSource: Record<string, string> = {
  site: "Сайт",
  website: "Сайт",
  whatsapp: "WhatsApp",
  "2gis": "2ГИС",
  instagram: "Instagram",
  manual: "Вручную",
};

export const apptSource: Record<string, string> = {
  manual: "Вручную",
  dashboard: "Дашборд",
  notebook: "Тетрадь",
  site: "Сайт",
  website: "Сайт",
  lead: "Заявка",
  api: "API",
};
