import type { Metadata } from "next";
import { ShieldCheck, ShieldAlert } from "lucide-react";

import { logoutAllAction } from "@/app/actions/auth";
import { SubmitButton } from "@/components/form";
import { ChangePasswordForm, TotpDisable, TotpEnroll } from "@/components/security";
import { PageHeader, Panel } from "@/components/ui";
import { ago } from "@/lib/format";
import { roles } from "@/lib/labels";
import { getMe } from "@/lib/me";

export const metadata: Metadata = { title: "Аккаунт и безопасность" };

export default async function AccountPage() {
  const { user } = await getMe();
  const isPlatform = user.role === "superadmin";

  return (
    <>
      <PageHeader
        title="Аккаунт и безопасность"
        description={`${user.name ? user.name + " · " : ""}${user.email} · ${roles[user.role]} · последний вход ${ago(user.last_login_at)}`}
      />
      <div className="flex max-w-3xl flex-col gap-6">
        <Panel
          title="Вход по коду (2FA)"
          description={
            isPlatform
              ? "Для аккаунтов Jazyl обязателен: открывает все клиники."
              : "Необязательно, но рекомендуем владельцам клиник: даже если пароль украдут, без телефона не войти."
          }
          actions={
            user.totp_enabled ? (
              <span className="inline-flex items-center gap-1.5 text-sm font-medium text-ok">
                <ShieldCheck className="size-4" aria-hidden /> Включён
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 text-sm font-medium text-action-ink">
                <ShieldAlert className="size-4" aria-hidden /> Выключен
              </span>
            )
          }
        >
          {user.totp_enabled ? (
            isPlatform ? (
              <p className="text-sm text-ink-2">
                Потеряли телефон? Код можно сбросить только на сервере: <code className="font-mono text-[13px]">jazyl reset-mfa {user.email}</code>
              </p>
            ) : (
              <TotpDisable />
            )
          ) : (
            <TotpEnroll />
          )}
        </Panel>

        <Panel title="Пароль" description="После смены пароля на других устройствах придётся войти заново.">
          <ChangePasswordForm />
        </Panel>

        <Panel title="Выйти на всех устройствах" description="Если вы входили с чужого компьютера или потеряли телефон.">
          <form action={logoutAllAction}>
            <SubmitButton variant="danger" pendingText="Выходим…">
              Выйти везде, включая это устройство
            </SubmitButton>
          </form>
        </Panel>
      </div>
    </>
  );
}
