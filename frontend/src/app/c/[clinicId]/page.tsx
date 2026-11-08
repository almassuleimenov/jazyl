import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight } from "lucide-react";

import { AppointmentsTable } from "@/components/appointments";
import { AutoRefresh } from "@/components/auto-refresh";
import { ButtonLink, Chip, Empty, Notice, PageHeader, Panel, StatStrip } from "@/components/ui";
import { api } from "@/lib/api";
import { dateTime, kzt, longToday, num, percent, phone } from "@/lib/format";
import { leadSource, waStatus } from "@/lib/labels";
import type { Appointment, Lead, Metrics, Page } from "@/lib/types";
import { getClinic } from "./data";

export const metadata: Metadata = { title: "Обзор" };

export default async function ClinicOverview({ params }: PageProps<"/c/[clinicId]">) {
  const { clinicId } = await params;
  const base = `/clinics/${encodeURIComponent(clinicId)}`;
  const [clinic, m, appts, leads] = await Promise.all([
    getClinic(clinicId),
    api<Metrics>(`${base}/analytics/summary`),
    api<{ items: Appointment[] }>(`${base}/appointments`),
    api<Page<Lead>>(`${base}/leads?status=new&limit=5`),
  ]);
  const tz = clinic.timezone;
  const wa = waStatus(clinic.wa_session_status);
  const replies = m.confirmed + m.reschedule_requested + m.cancelled_by_patient;

  return (
    <>
      <AutoRefresh seconds={60} />
      <PageHeader
        title="Обзор"
        description={longToday(tz)}
        actions={<ButtonLink href={`/c/${clinicId}/appointments?new=1`} variant="primary">Добавить запись</ButtonLink>}
      />

      {wa.tone !== "green" && clinic.wa_provider !== "mock" ? (
        <div className="mb-6">
          <Notice tone="amber">
            WhatsApp клиники {wa.label.toLowerCase()}: напоминания не уходят и ждут в очереди.{" "}
            <Link href={`/c/${clinicId}/settings#whatsapp`} className="font-medium underline">
              Подключить
            </Link>
          </Notice>
        </div>
      ) : null}

      <StatStrip
        stats={[
          { label: "Записей за 30 дней", value: num(m.appointments), hint: `${num(m.reminders_sent)} напоминаний отправлено` },
          {
            label: "Подтвердили",
            value: num(m.confirmed),
            hint: replies ? `${percent(m.confirmation_rate)} от ответивших` : "ответов пока нет",
          },
          { label: "Просили перенос", value: num(m.reschedule_requested), hint: `${num(m.cancelled_by_patient)} отменили сами` },
          {
            label: "Сохранено",
            value: kzt(m.saved_revenue_kzt),
            hint: `${num(m.prevented_no_shows)} предотвращённых неявок`,
            emphasis: true,
          },
        ]}
      />

      <div className="mt-6 grid grid-cols-[minmax(0,1fr)] gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <Panel
          title="Сегодня и завтра"
          description="Статусы обновляются, когда пациент отвечает 1 или 2 в WhatsApp."
          actions={
            <Link href={`/c/${clinicId}/appointments`} className="inline-flex items-center gap-1 text-sm font-medium text-brand hover:underline">
              Все записи <ArrowRight className="size-3.5" aria-hidden />
            </Link>
          }
          flush
        >
          <AppointmentsTable
            items={appts.items}
            clinicId={clinicId}
            tz={tz}
            emptyText={{
              title: "На сегодня и завтра записей нет",
              body: "Добавьте запись вручную или отправьте боту список на завтра с номера администратора.",
            }}
          />
        </Panel>

        <Panel
          title="Новые заявки"
          actions={
            leads.items.length ? (
              <Link href={`/c/${clinicId}/leads`} className="text-sm font-medium text-brand hover:underline">
                Все
              </Link>
            ) : null
          }
          flush
        >
          {leads.items.length === 0 ? (
            <Empty title="Новых заявок нет">Заявки с сайта и из WhatsApp появятся здесь.</Empty>
          ) : (
            <ul className="divide-y divide-line">
              {leads.items.map((l) => (
                <li key={l.id} className="px-5 py-3.5">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="text-sm font-medium break-words text-ink">{l.name || "Без имени"}</p>
                      <a href={`tel:+${l.phone}`} className="num text-xs text-ink-2 hover:text-brand hover:underline">
                        {phone(l.phone)}
                      </a>
                    </div>
                    <Chip tone="amber">{leadSource[l.source] ?? l.source}</Chip>
                  </div>
                  <p className="mt-1 text-xs text-ink-3">
                    {[l.service, dateTime(l.created_at, tz)].filter(Boolean).join(" · ")}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </Panel>
      </div>
    </>
  );
}
