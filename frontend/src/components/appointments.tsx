import { setAppointmentStatusAction } from "@/app/actions/clinic";
import { InlineAction } from "./form";
import { Chip, Empty, Table, Td, Th, Tr } from "./ui";
import { phone, relativeDay, time, ymd } from "@/lib/format";
import { apptSource, apptStatus } from "@/lib/labels";
import type { ApptStatus, Appointment } from "@/lib/types";

const next: Partial<
  Record<
    ApptStatus,
    { status: ApptStatus; label: string; variant?: "danger" | "secondary" }[]
  >
> = {
  scheduled: [
    { status: "confirmed", label: "Подтвердил" },
    { status: "cancelled", label: "Отмена", variant: "danger" },
  ],
  reschedule_requested: [
    { status: "confirmed", label: "Оставить время" },
    { status: "cancelled", label: "Отмена", variant: "danger" },
  ],
  confirmed: [
    { status: "completed", label: "Пришёл" },
    { status: "no_show", label: "Не пришёл", variant: "danger" },
  ],
};

export function AppointmentsTable({
  items,
  clinicId,
  tz,
  emptyText,
}: {
  items: Appointment[];
  clinicId: string;
  tz: string;
  emptyText: { title: string; body?: string };
}) {
  if (items.length === 0)
    return <Empty title={emptyText.title}>{emptyText.body}</Empty>;

  const days = new Map<string, Appointment[]>();
  for (const a of [...items].sort((x, y) =>
    x.starts_at.localeCompare(y.starts_at),
  )) {
    const d = ymd(new Date(a.starts_at), tz);
    days.set(d, [...(days.get(d) ?? []), a]);
  }

  const actions = (a: Appointment) =>
    (next[a.status] ?? []).map((n) => (
      <InlineAction
        key={n.status}
        action={setAppointmentStatusAction}
        fields={{ clinic_id: clinicId, appointment_id: a.id, status: n.status }}
        variant={n.variant ?? "secondary"}
        confirmText={
          n.status === "cancelled"
            ? "Отменить запись? Пациенту больше не придут напоминания."
            : undefined
        }
      >
        {n.label}
      </InlineAction>
    ));

  return (
    <>
      {/* Phones (owners between patients): one card per visit, big tap targets. */}
      <div className="md:hidden">
        {[...days.entries()].map(([day, list]) => (
          <section key={day} aria-label={relativeDay(day, tz)}>
            <h3 className="border-b border-line bg-sunk/60 px-4 py-2 font-sans text-[13px] font-semibold text-ink">
              {relativeDay(day, tz)}
              <span className="ml-2 font-normal text-ink-2">
                {list.length}{" "}
                {plural(list.length, "запись", "записи", "записей")}
              </span>
            </h3>
            <ul className="divide-y divide-line">
              {list.map((a) => {
                const st = apptStatus[a.status];
                return (
                  <li key={a.id} className="flex flex-col gap-2 px-4 py-3">
                    <div className="flex items-start gap-3">
                      <span className="num w-12 shrink-0 pt-px font-display text-[15px] font-bold text-ink">
                        {time(a.starts_at, tz)}
                      </span>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium break-words text-ink">
                          {a.patient_name || "Без имени"}
                        </p>
                        <p className="line-clamp-2 text-xs break-words text-ink-2">
                          {[a.service_name, a.doctor_name]
                            .filter(Boolean)
                            .join(" · ") || "—"}
                        </p>
                        <a
                          href={`tel:+${a.patient_phone}`}
                          className="num mt-0.5 inline-block py-1 text-[13px] text-brand"
                        >
                          {phone(a.patient_phone)}
                        </a>
                      </div>
                      <Chip tone={st.tone}>{st.label}</Chip>
                    </div>
                    {next[a.status] ? (
                      <div className="flex flex-wrap gap-1.5 pl-15">
                        {actions(a)}
                      </div>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          </section>
        ))}
      </div>
      <div className="hidden md:block">
        <Table label="Записи">
          <thead>
            <tr>
              <Th className="w-20">Время</Th>
              <Th>Пациент</Th>
              <Th>Статус</Th>
              <Th className="text-right">Действия</Th>
            </tr>
          </thead>
          {[...days.entries()].map(([day, list]) => {
            const waiting = list.filter((a) => a.status === "scheduled").length;
            return (
              <tbody key={day}>
                <tr>
                  <th
                    colSpan={4}
                    scope="colgroup"
                    className="border-b border-line bg-sunk/60 px-5 py-2 text-left text-[13px] font-semibold text-ink"
                  >
                    {relativeDay(day, tz)}
                    <span className="ml-2 font-normal text-ink-2">
                      {list.length}{" "}
                      {plural(list.length, "запись", "записи", "записей")}
                      {waiting ? ` · ${waiting} ждут ответа` : ""}
                    </span>
                  </th>
                </tr>
                {list.map((a) => {
                  const st = apptStatus[a.status];
                  return (
                    <Tr key={a.id}>
                      <Td className="num font-semibold whitespace-nowrap text-ink">
                        {time(a.starts_at, tz)}
                      </Td>
                      <Td>
                        <span className="block font-medium text-ink">
                          {a.patient_name || "Без имени"}
                        </span>
                        <span className="block text-xs text-ink-2">
                          <a
                            href={`tel:+${a.patient_phone}`}
                            className="num whitespace-nowrap hover:text-brand hover:underline"
                          >
                            {phone(a.patient_phone)}
                          </a>
                          {[a.service_name, a.doctor_name, a.source === "manual" ? "" : (apptSource[a.source] ?? a.source)]
                            .filter(Boolean)
                            .map((t) => ` · ${t}`)
                            .join("")}
                        </span>
                      </Td>
                      <Td>
                        <Chip tone={st.tone}>{st.label}</Chip>
                      </Td>
                      <Td>
                        <div className="flex flex-nowrap justify-end gap-1.5 whitespace-nowrap">
                          {actions(a)}
                        </div>
                      </Td>
                    </Tr>
                  );
                })}
              </tbody>
            );
          })}
        </Table>
      </div>
    </>
  );
}

export function plural(n: number, one: string, few: string, many: string) {
  const m10 = n % 10;
  const m100 = n % 100;
  if (m10 === 1 && m100 !== 11) return one;
  if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return few;
  return many;
}
