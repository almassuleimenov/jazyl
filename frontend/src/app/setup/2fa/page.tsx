import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { TotpEnroll } from "@/components/security";
import { Panel } from "@/components/ui";
import { getMe } from "@/lib/me";

export const metadata: Metadata = { title: "Двухфакторная защита" };

export default async function SetupTotpPage() {
  const me = await getMe();
  if (me.session.must_change_password) redirect("/setup/password");
  if (!me.session.mfa_setup_required) redirect("/");
  return (
    <>
      <h1 className="text-2xl font-bold text-ink">Включите вход по коду</h1>
      <p className="mt-2 mb-8 text-sm leading-relaxed text-ink-2">
        Аккаунт команды Jazyl открывает все клиники и их пациентов, поэтому кроме пароля нужен код с вашего телефона. Это
        займёт минуту.
      </p>
      <Panel>
        <TotpEnroll forced />
      </Panel>
    </>
  );
}
