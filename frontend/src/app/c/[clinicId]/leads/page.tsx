import type { Metadata } from "next";
import Link from "next/link";

import { convertLeadAction, updateLeadAction } from "@/app/actions/clinic";
import { ActionForm, InlineAction, SubmitButton } from "@/components/form";
import { Chip, Empty, Input, PageHeader, Panel, cx } from "@/components/ui";
import { api } from "@/lib/api";
import { addDays, dateTime, phone, ymd } from "@/lib/format";
import { leadSource, leadStatus } from "@/lib/labels";
import type { Lead, LeadStatus, Page } from "@/lib/types";
import { getClinic } from "../data";

export const metadata: Metadata = { title: "Заявки" };

const filters: { key: LeadStatus | "all"; label: string }[] = [
  { key: "new", label: "Новые" },
  { key: "contacted", label: "Связались" },
  { key: "booked", label: "Записаны" },
  { key: "rejected", label: "Отказ" },
  { key: "all", label: "Все" },
];

export default async function LeadsPage({ params, searchParams }: PageProps<"/c/[clinicId]/leads">) {
  const { clinicId } = await params;
  const sp = await searchParams;
  const raw = Array.isArray(sp.status) ? sp.status[0] : sp.status;
  const status = filters.find((f) => f.key === raw)?.key ?? "new";
  const clinic = await getClinic(clinicId);
  const tz = clinic.timezone;
  const q = status === "all" ? "" : `&status=${status}`;
  const leads = await api<Page<Lead>>(`/clinics/${encodeURIComponent(clinicId)}/leads?limit=100${q}`);
  const tomorrow = addDays(ymd(new Date(), tz), 1);

  return (
    <>
      <PageHeader title="Заявки" description="С сайта клиники и из WhatsApp. Позвоните пациенту и запишите его на время." />

      <nav aria-label="Статус" className="mb-4 flex w-fit max-w-full gap-1 overflow-x-auto rounded-control border border-line bg-surface p-1 whitespace-nowrap shadow-panel">
        {filters.map((f) => (
          <Link
            key={f.key}
            href={`?status=${f.key}`}
            aria-current={status === f.key ? "true" : undefined}
            className={cx(
              "rounded-[6px] px-3 py-1.5 text-sm transition-colors",
              status === f.key ? "bg-brand text-white" : "text-ink-2 hover:bg-sunk hover:text-ink",
            )}
          >
            {f.label}
          </Link>
        ))}
      </nav>

      <Panel flush>
        {leads.items.length === 0 ? (
          <Empty title={status === "new" ? "Новых заявок нет" : "Здесь пока пусто"}>
            {status === "new" ? "Как только пациент оставит заявку на сайте, она появится здесь." : null}
          </Empty>
        ) : (
          <ul className="divide-y divide-line">
            {leads.items.map((l) => {
              const st = leadStatus[l.status];
              const open = l.status === "new" || l.status === "contacted";
              return (
                <li key={l.id} className="grid grid-cols-[minmax(0,1fr)] gap-4 px-5 py-4 lg:grid-cols-[minmax(0,1fr)_auto]">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium text-ink">{l.name || "Без имени"}</p>
                      <Chip tone={st.tone}>{st.label}</Chip>
                      <span className="text-xs text-ink-3">
                        {leadSource[l.source] ?? l.source} · {dateTime(l.created_at, tz)}
                      </span>
                    </div>
                    <a href={`tel:+${l.phone}`} className="num mt-1 inline-block text-sm font-medium text-brand hover:underline">
                      {phone(l.phone)}
                    </a>
                    <dl className="mt-1.5 grid gap-x-6 gap-y-1 text-[13px] sm:grid-cols-2">
                      {l.service ? (
                        <div className="flex gap-1.5">
                          <dt className="text-ink-3">Услуга:</dt>
                          <dd className="text-ink-2">{l.service}</dd>
                        </div>
                      ) : null}
                      {l.preferred_at ? (
                        <div className="flex gap-1.5">
                          <dt className="text-ink-3">Удобно:</dt>
                          <dd className="text-ink-2">{dateTime(l.preferred_at, tz)}</dd>
                        </div>
                      ) : null}
                      {l.comment ? (
                        <div className="flex gap-1.5 sm:col-span-2">
                          <dt className="text-ink-3">Комментарий:</dt>
                          <dd className="text-ink-2">{l.comment}</dd>
                        </div>
                      ) : null}
                    </dl>
                  </div>

                  {open ? (
                    <div className="flex flex-col gap-2 lg:items-end">
                      <ActionForm action={convertLeadAction} className="flex-row! flex-wrap items-end gap-2!">
                        <input type="hidden" name="clinic_id" value={clinicId} />
                        <input type="hidden" name="lead_id" value={l.id} />
                        <label className="sr-only" htmlFor={`d-${l.id}`}>
                          Дата приёма
                        </label>
                        <Input id={`d-${l.id}`} name="date" type="date" required defaultValue={tomorrow} className="h-8 w-auto text-[13px]" />
                        <label className="sr-only" htmlFor={`t-${l.id}`}>
                          Время приёма
                        </label>
                        <Input id={`t-${l.id}`} name="time" type="time" required step={300} className="h-8 w-auto text-[13px]" />
                        <SubmitButton size="sm" variant="secondary" pendingText="Записываем…">
                          Записать
                        </SubmitButton>
                      </ActionForm>
                      <div className="flex gap-1.5">
                        {l.status === "new" ? (
                          <InlineAction action={updateLeadAction} fields={{ clinic_id: clinicId, lead_id: l.id, status: "contacted" }}>
                            Связались
                          </InlineAction>
                        ) : null}
                        <InlineAction
                          action={updateLeadAction}
                          fields={{ clinic_id: clinicId, lead_id: l.id, status: "rejected" }}
                          variant="ghost"
                        >
                          Отказ
                        </InlineAction>
                      </div>
                    </div>
                  ) : null}
                </li>
              );
            })}
          </ul>
        )}
      </Panel>
    </>
  );
}
