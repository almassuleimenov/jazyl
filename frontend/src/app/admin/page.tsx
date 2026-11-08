import type { Metadata } from "next";
import Link from "next/link";
import { Plus } from "lucide-react";

import { ButtonLink, Chip, Empty, PageHeader, Panel, StatStrip, Table, Td, Th, Tr } from "@/components/ui";
import { api } from "@/lib/api";
import { ago, kzt, num } from "@/lib/format";
import { clinicStatus, plans, waStatus } from "@/lib/labels";
import type { PlatformOverview } from "@/lib/types";

export const metadata: Metadata = { title: "Клиники" };

export default async function AdminOverview() {
  const o = await api<PlatformOverview>("/admin/overview");
  const pilots = o.clinics.filter((c) => c.status === "pilot").length;
  const waDown = o.clinics.filter((c) => (c.status === "active" || c.status === "pilot") && waStatus(c.wa_session_status).tone !== "green").length;
  const failed = o.clinics.reduce((s, c) => s + c.failed_30d, 0);

  return (
    <>
      <PageHeader
        title="Клиники"
        description="Все клиники Jazyl и их активность за 30 дней."
        actions={
          <ButtonLink href="/admin/clinics/new" variant="primary">
            <Plus className="size-4" aria-hidden />
            Новая клиника
          </ButtonLink>
        }
      />

      <StatStrip
        stats={[
          { label: "MRR", value: kzt(o.mrr_kzt), hint: `${o.paying} платящих`, emphasis: true },
          { label: "Клиник всего", value: num(o.total), hint: pilots ? `${pilots} на пилоте` : undefined },
          {
            label: "WhatsApp не подключён",
            value: num(waDown),
            hint: waDown ? "напоминания этих клиник ждут" : "у всех работающих клиник",
          },
          { label: "Сбои отправки, 30 дн", value: num(failed), hint: failed ? "проверьте клиники ниже" : "ошибок нет" },
        ]}
      />

      <Panel title="Все клиники" className="mt-6" flush>
        {o.clinics.length === 0 ? (
          <Empty
            title="Пока нет ни одной клиники"
            action={
              <ButtonLink href="/admin/clinics/new" variant="secondary">
                Добавить первую клинику
              </ButtonLink>
            }
          >
            Добавьте клинику, подключите её WhatsApp по QR-коду и выдайте владельцу логин.
          </Empty>
        ) : (
          <Table label="Клиники">
            <thead>
              <tr>
                <Th>Клиника</Th>
                <Th>Тариф</Th>
                <Th>Статус</Th>
                <Th>WhatsApp</Th>
                <Th className="text-right">Записей</Th>
                <Th className="text-right">Напоминаний</Th>
                <Th className="text-right">Сбоев</Th>
                <Th>Пациенты писали</Th>
              </tr>
            </thead>
            <tbody>
              {o.clinics.map((c) => {
                const wa = waStatus(c.wa_session_status);
                const st = clinicStatus[c.status];
                return (
                  <Tr key={c.clinic_id}>
                    <Td>
                      <Link href={`/c/${c.clinic_id}`} className="font-medium text-ink hover:text-brand hover:underline">
                        {c.name}
                      </Link>
                      {c.city ? <span className="block text-xs text-ink-3">{c.city}</span> : null}
                    </Td>
                    <Td className="whitespace-nowrap text-ink-2">{plans[c.plan]?.label ?? c.plan}</Td>
                    <Td>
                      <Chip tone={st.tone}>{st.label}</Chip>
                    </Td>
                    <Td>
                      <Chip tone={wa.tone}>{wa.label}</Chip>
                    </Td>
                    <Td className="num text-right">{num(c.appointments_30d)}</Td>
                    <Td className="num text-right">{num(c.reminders_30d)}</Td>
                    <Td className={`num text-right ${c.failed_30d ? "font-semibold text-bad" : "text-ink-3"}`}>{num(c.failed_30d)}</Td>
                    <Td className="whitespace-nowrap text-ink-2">{ago(c.last_inbound_at)}</Td>
                  </Tr>
                );
              })}
            </tbody>
          </Table>
        )}
      </Panel>
    </>
  );
}
