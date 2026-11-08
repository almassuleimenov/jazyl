import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { revokeSiteKeyAction } from "@/app/actions/clinic";
import { InlineAction } from "@/components/form";
import { SiteKeyCreate } from "@/components/site-key-create";
import { Chip, Empty, Notice, PageHeader, Panel, Table, Td, Th, Tr } from "@/components/ui";
import { api } from "@/lib/api";
import { ago, dateTime } from "@/lib/format";
import { plans } from "@/lib/labels";
import { getMe } from "@/lib/me";
import type { SiteKey } from "@/lib/types";
import { getClinic } from "../data";

export const metadata: Metadata = { title: "Сайт" };

export default async function SitePage({ params }: PageProps<"/c/[clinicId]/site">) {
  const { clinicId } = await params;
  const me = await getMe();
  if (me.user.role === "admin") notFound();
  const clinic = await getClinic(clinicId);
  const allowed = clinic.plan === "full_chair" || clinic.plan === "growth";
  const keys = await api<{ items: SiteKey[] }>(`/clinics/${encodeURIComponent(clinicId)}/site-keys`);
  const tz = clinic.timezone;

  return (
    <>
      <PageHeader
        title="Сайт"
        description="Ключ подключает сайт клиники к Jazyl: заявки с формы записи попадают в «Заявки», а свободное время берётся отсюда."
      />
      <div className="flex max-w-4xl flex-col gap-6">
        {!allowed ? (
          <Notice tone="amber">
            Приём заявок с сайта входит в тарифы «{plans.full_chair.label}» и «{plans.growth.label}». Сейчас у клиники «
            {plans[clinic.plan].label}».
          </Notice>
        ) : null}

        <Panel title="Ключи сайта" flush>
          {keys.items.length === 0 ? (
            <Empty title="Ключей пока нет">Создайте ключ и передайте его разработчику сайта.</Empty>
          ) : (
            <Table label="Ключи сайта">
              <thead>
                <tr>
                  <Th>Название</Th>
                  <Th>Ключ</Th>
                  <Th>Домены</Th>
                  <Th>Использован</Th>
                  <Th className="text-right" />
                </tr>
              </thead>
              <tbody>
                {keys.items.map((k) => (
                  <Tr key={k.id} className={k.revoked_at ? "opacity-60" : ""}>
                    <Td className="font-medium text-ink">
                      {k.name || "Без названия"}
                      <span className="block text-xs font-normal text-ink-3">создан {dateTime(k.created_at, tz)}</span>
                    </Td>
                    <Td>
                      <code className="font-mono text-[13px] text-ink-2">{k.key_prefix}…</code>
                    </Td>
                    <Td className="text-[13px] text-ink-2">{k.allowed_origins.length ? k.allowed_origins.join(", ") : "только с сервера"}</Td>
                    <Td className="whitespace-nowrap text-ink-2">{ago(k.last_used_at)}</Td>
                    <Td className="text-right">
                      {k.revoked_at ? (
                        <Chip tone="neutral">Отозван</Chip>
                      ) : (
                        <InlineAction
                          action={revokeSiteKeyAction}
                          fields={{ clinic_id: clinicId, key_id: k.id }}
                          variant="danger"
                          confirmText="Отозвать ключ? Форма записи на сайте перестанет работать, пока не вставите новый."
                        >
                          Отозвать
                        </InlineAction>
                      )}
                    </Td>
                  </Tr>
                ))}
              </tbody>
            </Table>
          )}
        </Panel>

        {allowed ? (
          <Panel title="Новый ключ" description="Ключ показывается один раз. Храните его в переменных окружения сайта (например, на Vercel).">
            <SiteKeyCreate clinicId={clinicId} />
          </Panel>
        ) : null}
      </div>
    </>
  );
}
