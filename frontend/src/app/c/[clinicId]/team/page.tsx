import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { KeyRound, ShieldCheck, X } from "lucide-react";

import {
  createUserAction,
  resetUserPasswordAction,
  resetUserTotpAction,
  setUserActiveAction,
} from "@/app/actions/clinic";
import { ActionForm, InlineAction, SubmitButton } from "@/components/form";
import { TempPassword } from "@/components/temp-password";
import {
  Chip,
  Empty,
  Field,
  Input,
  Notice,
  PageHeader,
  Panel,
  Select,
  Table,
  Td,
  Th,
  Tr,
  buttonClass,
  cx,
} from "@/components/ui";
import { api } from "@/lib/api";
import { ago } from "@/lib/format";
import { roles } from "@/lib/labels";
import { getMe } from "@/lib/me";
import type { User } from "@/lib/types";

export const metadata: Metadata = { title: "Команда и доступ" };

export default async function TeamPage({
  params,
  searchParams,
}: PageProps<"/c/[clinicId]/team">) {
  const { clinicId } = await params;
  const sp = await searchParams;
  const me = await getMe();
  if (me.user.role === "admin") notFound();
  const isPlatform = me.user.role === "superadmin";
  const users = await api<{ items: User[] }>(
    `/clinics/${encodeURIComponent(clinicId)}/users`,
  );
  const justCreated = sp.created === "1";

  // Owners may manage administrators; only Jazyl manages owners.
  const canManage = (u: User) =>
    u.id !== me.user.id && (isPlatform || u.role === "admin");

  return (
    <>
      <PageHeader
        title="Команда и доступ"
        description="Каждый сотрудник входит под своей почтой. Пароль, который вы выдаёте, временный: при первом входе его нужно сменить."
      />

      {justCreated ? (
        <div className="mb-6">
          <Notice tone="green">
            Клиника создана. Теперь выдайте доступ владельцу ниже, затем{" "}
            <Link
              href={`/c/${clinicId}/settings#whatsapp`}
              className="font-medium underline"
            >
              подключите WhatsApp клиники
            </Link>
            .
          </Notice>
        </div>
      ) : null}

      <div className="flex flex-col gap-6">
        <Panel flush>
          {users.items.length === 0 ? (
            <Empty title="В клинике пока нет сотрудников">
              Создайте первый доступ — обычно это владелец клиники.
            </Empty>
          ) : (
            <>
              {/* Phone: one card per person, actions full width. */}
              <ul
                className="divide-y divide-line md:hidden"
                aria-label="Сотрудники"
              >
                {users.items.map((u) => (
                  <li
                    key={u.id}
                    className={cx(
                      "flex flex-col gap-3 px-4 py-4",
                      !u.is_active && "opacity-60",
                    )}
                  >
                    <div className="min-w-0">
                      <Person u={u} self={u.id === me.user.id} />
                      <p className="mt-1 text-xs text-ink-2">
                        {roles[u.role]} · вход {ago(u.last_login_at)}
                      </p>
                    </div>
                    <Protection u={u} />
                    {canManage(u) ? (
                      <UserActions
                        u={u}
                        clinicId={clinicId}
                        isPlatform={isPlatform}
                        wide
                      />
                    ) : null}
                  </li>
                ))}
              </ul>

              <div className="hidden md:block">
                <Table label="Сотрудники">
                  <thead>
                    <tr>
                      <Th>Сотрудник</Th>
                      <Th>Роль</Th>
                      <Th>Защита</Th>
                      <Th>Последний вход</Th>
                      <Th className="text-right">Доступ</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.items.map((u) => (
                      <Tr
                        key={u.id}
                        className={u.is_active ? "" : "opacity-60"}
                      >
                        <Td>
                          <Person u={u} self={u.id === me.user.id} />
                        </Td>
                        <Td className="text-ink-2">{roles[u.role]}</Td>
                        <Td>
                          <Protection u={u} />
                        </Td>
                        <Td className="whitespace-nowrap text-ink-2">
                          {ago(u.last_login_at)}
                        </Td>
                        <Td>
                          {canManage(u) ? (
                            <UserActions
                              u={u}
                              clinicId={clinicId}
                              isPlatform={isPlatform}
                            />
                          ) : null}
                        </Td>
                      </Tr>
                    ))}
                  </tbody>
                </Table>
              </div>

              {/* Rendered once, outside the scrolling table, in the top layer. */}
              {users.items.filter(canManage).map((u) => (
                <ResetPassword key={u.id} u={u} clinicId={clinicId} />
              ))}
            </>
          )}
        </Panel>

        <Panel
          title="Выдать доступ"
          description="Передайте почту и пароль лично или звонком, не пересылайте пароль в общих чатах."
          className="max-w-3xl"
        >
          <ActionForm action={createUserAction} resetOnSuccess>
            <input type="hidden" name="clinic_id" value={clinicId} />
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Почта" htmlFor="email">
                <Input
                  id="email"
                  name="email"
                  type="email"
                  required
                  autoComplete="off"
                  placeholder="owner@clinic.kz"
                />
              </Field>
              <Field label="Имя" htmlFor="name">
                <Input
                  id="name"
                  name="name"
                  autoComplete="off"
                  placeholder="Айгуль Серикова"
                />
              </Field>
              <Field
                label="Роль"
                htmlFor="role"
                hint={
                  isPlatform
                    ? "Владелец видит цифры и управляет командой; администратор работает с записями."
                    : "Владельца добавляет команда Jazyl."
                }
              >
                <Select
                  id="role"
                  name="role"
                  defaultValue={
                    isPlatform && users.items.length === 0 ? "owner" : "admin"
                  }
                >
                  {isPlatform ? <option value="owner">Владелец</option> : null}
                  <option value="admin">Администратор</option>
                </Select>
              </Field>
              <Field label="Временный пароль" htmlFor="new-password">
                <TempPassword id="new-password" />
              </Field>
            </div>
            <div>
              <SubmitButton variant="primary" pendingText="Создаём…">
                Создать доступ
              </SubmitButton>
            </div>
          </ActionForm>
        </Panel>
      </div>
    </>
  );
}

function Person({ u, self }: { u: User; self: boolean }) {
  return (
    <>
      <span className="block font-medium break-words text-ink">
        {u.name || "Без имени"}
        {self ? (
          <span className="ml-1.5 text-xs font-normal text-ink-3">это вы</span>
        ) : null}
      </span>
      <span className="block text-xs break-all text-ink-2">{u.email}</span>
    </>
  );
}

function Protection({ u }: { u: User }) {
  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
      {!u.is_active ? (
        <Chip tone="neutral">Отключён</Chip>
      ) : u.must_change_password ? (
        <Chip tone="amber">Ждёт смены пароля</Chip>
      ) : null}
      {u.totp_enabled ? (
        <span className="inline-flex items-center gap-1 text-xs font-medium text-ok">
          <ShieldCheck className="size-3.5" aria-hidden /> 2FA
        </span>
      ) : (
        <span className="text-xs text-ink-3">Только пароль</span>
      )}
    </div>
  );
}

const resetId = (u: User) => `reset-${u.id}`;

function UserActions({
  u,
  clinicId,
  isPlatform,
  wide = false,
}: {
  u: User;
  clinicId: string;
  isPlatform: boolean;
  wide?: boolean;
}) {
  return (
    <div
      className={cx(
        "flex flex-wrap gap-1.5",
        wide ? "[&>*]:flex-1 [&_button]:w-full" : "flex-nowrap justify-end",
      )}
    >
      <button
        type="button"
        popoverTarget={resetId(u)}
        className={buttonClass("secondary", "sm", "whitespace-nowrap")}
      >
        <KeyRound className="size-3.5" aria-hidden />
        Новый пароль
      </button>
      {isPlatform && u.totp_enabled ? (
        <InlineAction
          action={resetUserTotpAction}
          fields={{ clinic_id: clinicId, user_id: u.id }}
          confirmText="Снять 2FA? Делайте это, только если сотрудник потерял телефон и вы проверили, что это он."
        >
          Снять 2FA
        </InlineAction>
      ) : null}
      <InlineAction
        action={setUserActiveAction}
        fields={{
          clinic_id: clinicId,
          user_id: u.id,
          active: u.is_active ? "0" : "1",
        }}
        variant={u.is_active ? "danger" : "secondary"}
        confirmText={
          u.is_active
            ? `Отключить доступ ${u.email}? Сотрудник сразу выйдет на всех устройствах.`
            : undefined
        }
      >
        {u.is_active ? "Отключить" : "Включить"}
      </InlineAction>
    </div>
  );
}

function ResetPassword({ u, clinicId }: { u: User; clinicId: string }) {
  return (
    <div
      id={resetId(u)}
      popover="auto"
      role="dialog"
      aria-label={`Новый пароль для ${u.email}`}
      className="m-auto w-[min(24rem,calc(100vw-2rem))] rounded-panel border border-line bg-surface p-5 text-ink shadow-pop backdrop:bg-ink/30"
    >
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-base font-semibold">Новый временный пароль</p>
          <p className="mt-0.5 text-xs break-all text-ink-2">{u.email}</p>
        </div>
        <button
          type="button"
          popoverTarget={resetId(u)}
          popoverTargetAction="hide"
          aria-label="Закрыть"
          className="grid size-8 shrink-0 place-items-center rounded-control text-ink-2 hover:bg-sunk hover:text-ink"
        >
          <X className="size-4" aria-hidden />
        </button>
      </div>
      <ActionForm action={resetUserPasswordAction}>
        <input type="hidden" name="clinic_id" value={clinicId} />
        <input type="hidden" name="user_id" value={u.id} />
        <Field
          label="Временный пароль"
          htmlFor={`pw-${u.id}`}
          hint="Все сессии сотрудника закроются, блокировка за неверные попытки снимется."
        >
          <TempPassword id={`pw-${u.id}`} />
        </Field>
        <div>
          <SubmitButton variant="secondary" pendingText="Сохраняем…">
            Выдать пароль
          </SubmitButton>
        </div>
      </ActionForm>
    </div>
  );
}
