// Formatting in the clinic's time zone and Russian locale.

export const DEFAULT_TZ = "Asia/Almaty";

const tenge = new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 0 });

export function kzt(n: number): string {
  return `${tenge.format(n)} ₸`;
}

export function num(n: number): string {
  return tenge.format(n);
}

export function percent(x: number): string {
  if (!Number.isFinite(x)) return "—";
  return `${Math.round(x * 100)}%`;
}

export function time(iso: string, tz = DEFAULT_TZ): string {
  return new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit", timeZone: tz }).format(new Date(iso));
}

export function dayLabel(iso: string, tz = DEFAULT_TZ): string {
  return new Intl.DateTimeFormat("ru-RU", { weekday: "short", day: "numeric", month: "long", timeZone: tz }).format(new Date(iso));
}

export function dateTime(iso: string | null | undefined, tz = DEFAULT_TZ): string {
  if (!iso) return "—";
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: tz,
  }).format(new Date(iso));
}

/** YYYY-MM-DD of `d` in time zone `tz`. */
export function ymd(d: Date, tz = DEFAULT_TZ): string {
  return new Intl.DateTimeFormat("en-CA", { year: "numeric", month: "2-digit", day: "2-digit", timeZone: tz }).format(d);
}

export function addDays(ymdStr: string, days: number): string {
  const [y, m, d] = ymdStr.split("-").map(Number);
  const t = new Date(Date.UTC(y, m - 1, d + days));
  return t.toISOString().slice(0, 10);
}

export function longToday(tz = DEFAULT_TZ): string {
  const s = new Intl.DateTimeFormat("ru-RU", { weekday: "long", day: "numeric", month: "long", timeZone: tz }).format(new Date());
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/** "Сегодня" / "Завтра" / "пт, 14 ноября" for a YYYY-MM-DD in the clinic zone. */
export function relativeDay(day: string, tz = DEFAULT_TZ): string {
  const today = ymd(new Date(), tz);
  if (day === today) return "Сегодня";
  if (day === addDays(today, 1)) return "Завтра";
  if (day === addDays(today, -1)) return "Вчера";
  const [y, m, d] = day.split("-").map(Number);
  return new Intl.DateTimeFormat("ru-RU", { weekday: "short", day: "numeric", month: "long", timeZone: "UTC" }).format(
    new Date(Date.UTC(y, m - 1, d)),
  );
}

/** 77011234567 → +7 701 123 45 67 */
export function phone(p: string | null | undefined): string {
  if (!p) return "—";
  const d = p.replace(/\D/g, "");
  if (d.length === 11 && d.startsWith("7")) {
    return `+7 ${d.slice(1, 4)} ${d.slice(4, 7)} ${d.slice(7, 9)} ${d.slice(9)}`;
  }
  return `+${d}`;
}

export function ago(iso: string | null | undefined): string {
  if (!iso) return "никогда";
  const s = Math.round((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return "только что";
  const m = Math.round(s / 60);
  if (m < 60) return `${m} мин назад`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h} ч назад`;
  const d = Math.round(h / 24);
  return `${d} дн назад`;
}

/** Datetime-local input value ("2026-11-09T10:30") in the clinic zone → RFC3339 with +05:00. */
export function localInputToISO(v: string, offset = "+05:00"): string {
  return `${v}:00${offset}`;
}
