import type { Metadata } from "next";
import Link from "next/link";

import { sendMessageAction } from "@/app/actions/clinic";
import { AutoRefresh } from "@/components/auto-refresh";
import { ActionForm, SubmitButton } from "@/components/form";
import {
  Empty,
  Notice,
  PageHeader,
  Panel,
  Textarea,
  cx,
} from "@/components/ui";
import { api } from "@/lib/api";
import { dateTime, phone } from "@/lib/format";
import type { ChatSummary, Message, Page } from "@/lib/types";
import { getClinic } from "../data";

export const metadata: Metadata = { title: "Переписка" };

const kindLabel: Record<string, string> = {
  reminder: "напоминание",
  review_request: "запрос оценки",
  review_link: "ссылка на 2ГИС",
  auto_reply: "автоответ",
  manual: "вручную",
  report: "отчёт",
};

export default async function ChatsPage({
  params,
  searchParams,
}: PageProps<"/c/[clinicId]/chats">) {
  const { clinicId } = await params;
  const sp = await searchParams;
  const selected =
    (Array.isArray(sp.phone) ? sp.phone[0] : sp.phone)?.replace(/\D/g, "") ||
    undefined;
  const base = `/clinics/${encodeURIComponent(clinicId)}`;
  const [clinic, chats, thread] = await Promise.all([
    getClinic(clinicId),
    api<{ items: ChatSummary[] }>(`${base}/chats`),
    selected
      ? api<Page<Message>>(`${base}/messages?phone=${selected}&limit=100`)
      : Promise.resolve(null),
  ]);
  const tz = clinic.timezone;
  const current = chats.items.find((c) => c.phone === selected);
  const messages = thread ? [...thread.items].reverse() : [];

  return (
    <>
      <AutoRefresh seconds={30} />
      <PageHeader
        title="Переписка"
        description="Кто писал боту за последние 30 дней. Бот понимает только 1, 2, оценку и «СТОП»; на остальное отвечаете вы."
      />
      {chats.items.length === 0 && !selected ? (
        <Panel>
          <Empty title="Сообщений пока нет">Когда пациенты ответят на напоминания, переписка появится здесь.</Empty>
        </Panel>
      ) : (
      <div className="grid grid-cols-[minmax(0,1fr)] gap-6 lg:grid-cols-[340px_minmax(0,1fr)]">
        <Panel flush className={cx(selected && "hidden lg:block")}>
          {chats.items.length === 0 ? (
            <Empty title="Сообщений пока нет">
              Когда пациенты ответят на напоминания, переписка появится здесь.
            </Empty>
          ) : (
            <ul className="max-h-[70dvh] divide-y divide-line overflow-y-auto">
              {chats.items.map((c) => (
                <li key={c.phone}>
                  <Link
                    href={`?phone=${c.phone}`}
                    aria-current={c.phone === selected ? "true" : undefined}
                    className={cx(
                      "block px-4 py-3 transition-colors hover:bg-[#fbfaf7]",
                      c.phone === selected &&
                        "bg-brand-soft hover:bg-brand-soft",
                    )}
                  >
                    <span className="flex items-baseline justify-between gap-2">
                      <span className="truncate text-sm font-medium text-ink">
                        {c.patient_name || phone(c.phone)}
                      </span>
                      <span className="num shrink-0 text-xs text-ink-3">
                        {dateTime(c.last_message_at, tz)}
                      </span>
                    </span>
                    <span className="mt-0.5 block truncate text-[13px] text-ink-2">
                      {c.last_direction === "out" ? "Бот: " : ""}
                      {c.last_body}
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </Panel>

        <Panel
          flush
          className={cx(!selected && "hidden lg:block")}
          title={
            selected
              ? current?.patient_name || phone(selected)
              : "Выберите переписку"
          }
          description={
            selected && current?.patient_name ? phone(selected) : undefined
          }
          actions={
            selected ? (
              <Link
                href="?"
                className="text-sm text-brand hover:underline lg:hidden"
              >
                ← Все
              </Link>
            ) : null
          }
        >
          {!selected ? (
            <Empty title="Переписка не выбрана">
              Откройте номер слева, чтобы увидеть сообщения.
            </Empty>
          ) : (
            <>
              <ol className="flex max-h-[60dvh] flex-col gap-2 overflow-y-auto bg-[#fbfaf7] px-4 py-4">
                {messages.length === 0 ? (
                  <li className="text-sm text-ink-2">Сообщений нет.</li>
                ) : null}
                {messages.map((m) => (
                  <li
                    key={m.id}
                    className={cx(
                      "flex",
                      m.direction === "out" ? "justify-end" : "justify-start",
                    )}
                  >
                    <div
                      className={cx(
                        "max-w-[80%] rounded-[12px] px-3 py-2 text-sm leading-relaxed whitespace-pre-wrap",
                        m.direction === "out"
                          ? "rounded-br-[4px] bg-brand-soft text-ink"
                          : "rounded-bl-[4px] border border-line bg-surface text-ink",
                      )}
                    >
                      {m.body}
                      <span className="mt-1 block text-[11px] text-ink-3">
                        {dateTime(m.sent_at ?? m.created_at, tz)}
                        {m.direction === "out" && kindLabel[m.kind]
                          ? ` · ${kindLabel[m.kind]}`
                          : ""}
                        {m.status === "failed"
                          ? " · не доставлено"
                          : m.status === "queued" || m.status === "pending"
                            ? " · в очереди"
                            : ""}
                      </span>
                    </div>
                  </li>
                ))}
              </ol>
              <div className="border-t border-line p-4">
                <ActionForm action={sendMessageAction} resetOnSuccess>
                  <input type="hidden" name="clinic_id" value={clinicId} />
                  <input type="hidden" name="phone" value={selected} />
                  <label htmlFor="text" className="sr-only">
                    Сообщение
                  </label>
                  <Textarea
                    id="text"
                    name="text"
                    rows={2}
                    maxLength={4000}
                    required
                    placeholder="Ответ пациенту от имени клиники"
                  />
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="text-xs text-ink-3">
                      Уйдёт с номера клиники с паузой против блокировки, обычно
                      в течение минуты.
                    </p>
                    <SubmitButton size="sm" pendingText="Отправляем…">
                      Отправить
                    </SubmitButton>
                  </div>
                </ActionForm>
              </div>
            </>
          )}
        </Panel>
      </div>
      )}
      {clinic.wa_provider === "mock" ? (
        <div className="mt-6">
          <Notice>
            Клиника в тестовом режиме: сообщения не уходят пациентам, только
            пишутся в журнал.
          </Notice>
        </div>
      ) : null}
    </>
  );
}
