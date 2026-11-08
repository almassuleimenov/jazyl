import type { Metadata } from "next";

import { Logo } from "@/components/logo";
import { safeNext } from "@/lib/session";
import { LoginForm } from "./login-form";

export const metadata: Metadata = { title: "Вход" };

export default async function LoginPage({ searchParams }: PageProps<"/login">) {
  const sp = await searchParams;
  const one = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v);
  const notice =
    one(sp.signed_out) === "all"
      ? "Вы вышли на всех устройствах."
      : one(sp.expired)
        ? "Сессия закончилась. Войдите снова."
        : undefined;

  return (
    <div className="flex min-h-dvh flex-col">
      <div className="flex flex-1 items-center justify-center px-4 py-12">
        <div className="w-full max-w-[400px]">
          <div className="mb-8 flex flex-col items-center gap-3 text-center">
            <Logo tone="dark" className="text-[30px]" />
            <h1 className="text-xl font-bold text-ink">Вход в кабинет</h1>
            <p className="text-sm text-ink-2">Для клиник-клиентов Jazyl и нашей команды</p>
          </div>
          <div className="rounded-panel border border-line bg-surface p-6 shadow-panel sm:p-7">
            <LoginForm next={safeNext(one(sp.next)) ?? ""} notice={notice} />
          </div>
          <p className="mt-6 text-center text-[13px] leading-relaxed text-ink-2">
            Доступ выдаёт команда Jazyl. Забыли пароль — напишите нам в WhatsApp, мы выдадим новый.
          </p>
        </div>
      </div>
    </div>
  );
}
