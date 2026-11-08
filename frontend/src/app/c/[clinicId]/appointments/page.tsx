import type { Metadata } from "next";
import Link from "next/link";
import { ChevronDown } from "lucide-react";

import { createAppointmentAction, importScheduleAction } from "@/app/actions/clinic";
import { AppointmentsTable } from "@/components/appointments";
import { AutoRefresh } from "@/components/auto-refresh";
import { ActionForm, SubmitButton } from "@/components/form";
import { Field, Input, PageHeader, Panel, Select, Textarea, cx } from "@/components/ui";
import { api } from "@/lib/api";
import { addDays, relativeDay, ymd } from "@/lib/format";
import type { Appointment } from "@/lib/types";
import { getClinic } from "../data";

export const metadata: Metadata = { title: "Записи" };

const ranges = [
  { key: "today", label: "Сегодня", offset: 0, days: 1 },
  { key: "tomorrow", label: "Завтра", offset: 1, days: 1 },
  { key: "week", label: "7 дней", offset: 0, days: 7 },
  { key: "past", label: "Прошлые 7 дней", offset: -7, days: 7 },
] as const;

export default async function AppointmentsPage({ params, searchParams }: PageProps<"/c/[clinicId]/appointments">) {
  const { clinicId } = await params;
  const sp = await searchParams;
  const one = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);
  const clinic = await getClinic(clinicId);
  const tz = clinic.timezone;
  const today = ymd(new Date(), tz);

  const day = one(sp.day);
  const range = ranges.find((r) => r.key === one(sp.range)) ?? (day ? undefined : ranges[1]);
  const from = day && /^\d{4}-\d{2}-\d{2}$/.test(day) ? day : addDays(today, range?.offset ?? 0);
  const to = addDays(from, range?.days ?? 1);

  const appts = await api<{ items: Appointment[] }>(
    `/clinics/${encodeURIComponent(clinicId)}/appointments?from=${from}&to=${to}`,
  );
  const openNew = one(sp.new) === "1";

  return (
    <>
      <AutoRefresh seconds={60} />
      <PageHeader title="Записи" description="Напоминания уходят накануне вечером и за 2 часа до приёма." />

      <div className="mb-4 flex flex-wrap items-center gap-2">
        <nav aria-label="Период" className="flex max-w-full gap-1 overflow-x-auto rounded-control border border-line bg-surface p-1 whitespace-nowrap shadow-panel">
          {ranges.map((r) => {
            const active = !day && range?.key === r.key;
            return (
              <Link
                key={r.key}
                href={`?range=${r.key}`}
                aria-current={active ? "true" : undefined}
                className={cx(
                  "rounded-[6px] px-3 py-1.5 text-sm transition-colors",
                  active ? "bg-brand text-white" : "text-ink-2 hover:bg-sunk hover:text-ink",
                )}
              >
                {r.label}
              </Link>
            );
          })}
        </nav>
        <form className="flex items-center gap-2" action="">
          <label htmlFor="day" className="sr-only">
            Дата
          </label>
          <Input id="day" name="day" type="date" defaultValue={day ?? ""} className="h-9 w-auto" />
          <button type="submit" className="h-9 rounded-control border border-line-strong bg-surface px-3 text-sm text-ink hover:bg-sunk">
            Показать
          </button>
        </form>
        {day ? <span className="text-sm text-ink-2">{relativeDay(from, tz)}</span> : null}
      </div>

      <div className="flex flex-col gap-6">
        <Panel flush>
          <AppointmentsTable
            items={appts.items}
            clinicId={clinicId}
            tz={tz}
            emptyText={{ title: "За этот период записей нет", body: "Добавьте запись ниже или выберите другой день." }}
          />
        </Panel>

        <div className="grid grid-cols-[minmax(0,1fr)] gap-6 lg:grid-cols-2">
          <details open={openNew} className="group rounded-panel border border-line bg-surface shadow-panel">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-5 py-4 [&::-webkit-details-marker]:hidden">
              <span>
                <span className="block font-display text-[15px] font-semibold text-ink">Новая запись</span>
                <span className="mt-0.5 block text-[13px] text-ink-2">Пациент получит напоминания автоматически.</span>
              </span>
              <ChevronDown className="size-4 text-ink-3 transition-transform group-open:rotate-180" aria-hidden />
            </summary>
            <div className="border-t border-line p-5">
              <ActionForm action={createAppointmentAction} resetOnSuccess>
                <input type="hidden" name="clinic_id" value={clinicId} />
                <div className="grid gap-4 sm:grid-cols-2">
                  <Field label="Телефон пациента" htmlFor="patient_phone">
                    <Input id="patient_phone" name="patient_phone" type="tel" inputMode="tel" required placeholder="+7 701 123 45 67" />
                  </Field>
                  <Field label="Имя" htmlFor="patient_name">
                    <Input id="patient_name" name="patient_name" autoComplete="off" placeholder="Айгерим" />
                  </Field>
                  <Field label="Дата" htmlFor="date">
                    <Input id="date" name="date" type="date" required defaultValue={addDays(today, 1)} min={today} />
                  </Field>
                  <Field label="Время" htmlFor="time">
                    <Input id="time" name="time" type="time" required step={300} defaultValue="10:00" />
                  </Field>
                  <Field label="Услуга" htmlFor="service_name">
                    <Input id="service_name" name="service_name" placeholder="Чистка" />
                  </Field>
                  <Field label="Длительность" htmlFor="duration_min">
                    <Select id="duration_min" name="duration_min" defaultValue="30">
                      {[15, 30, 45, 60, 90, 120].map((m) => (
                        <option key={m} value={m}>
                          {m} мин
                        </option>
                      ))}
                    </Select>
                  </Field>
                  <Field label="Язык сообщений" htmlFor="patient_language">
                    <Select id="patient_language" name="patient_language" defaultValue={clinic.language || "ru"}>
                      <option value="ru">Русский</option>
                      <option value="kk">Қазақша</option>
                    </Select>
                  </Field>
                </div>
                <div>
                  <SubmitButton variant="primary" pendingText="Сохраняем…">
                    Добавить запись
                  </SubmitButton>
                </div>
              </ActionForm>
            </div>
          </details>

          <details className="group rounded-panel border border-line bg-surface shadow-panel">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-5 py-4 [&::-webkit-details-marker]:hidden">
              <span>
                <span className="block font-display text-[15px] font-semibold text-ink">Список из тетради</span>
                <span className="mt-0.5 block text-[13px] text-ink-2">Вставьте записи текстом, по одной в строке.</span>
              </span>
              <ChevronDown className="size-4 text-ink-3 transition-transform group-open:rotate-180" aria-hidden />
            </summary>
            <div className="border-t border-line p-5">
              <ActionForm action={importScheduleAction} resetOnSuccess>
                <input type="hidden" name="clinic_id" value={clinicId} />
                <Field label="Записи" htmlFor="text" hint="Тот же формат, что и в WhatsApp-боте. Первая строка может быть датой.">
                  <Textarea
                    id="text"
                    name="text"
                    rows={7}
                    required
                    className="font-mono text-[13px]"
                    placeholder={"завтра\n10:00 Айгерим 87011234567 чистка\n11:30 Ержан 87029876543 пломба"}
                  />
                </Field>
                <div>
                  <SubmitButton variant="secondary" pendingText="Разбираем…">Загрузить список</SubmitButton>
                </div>
              </ActionForm>
            </div>
          </details>
        </div>
      </div>
    </>
  );
}
