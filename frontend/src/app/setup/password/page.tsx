import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { ChangePasswordForm } from "@/components/security";
import { Panel } from "@/components/ui";
import { getMe } from "@/lib/me";

export const metadata: Metadata = { title: "Новый пароль" };

export default async function SetupPasswordPage() {
  const me = await getMe();
  if (!me.session.must_change_password) redirect("/");
  return (
    <>
      <h1 className="text-2xl font-bold text-ink">Придумайте свой пароль</h1>
      <p className="mt-2 mb-8 text-sm leading-relaxed text-ink-2">
        {me.user.name ? `${me.user.name}, в` : "В"}ы вошли с временным паролем, который выдала команда Jazyl. Замените его на
        свой: так доступ к записям пациентов будет только у вас.
      </p>
      <Panel>
        <ChangePasswordForm forced />
      </Panel>
    </>
  );
}
