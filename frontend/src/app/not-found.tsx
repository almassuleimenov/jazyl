import Link from "next/link";

import { Logo } from "@/components/logo";
import { buttonClass } from "@/components/ui";

export default function NotFound() {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-4 px-4 text-center">
      <Logo tone="dark" className="text-[28px]" />
      <h1 className="text-xl font-bold text-ink">Страница не найдена</h1>
      <p className="max-w-sm text-sm text-ink-2">Её нет, или у вашего аккаунта нет к ней доступа.</p>
      <Link href="/" className={buttonClass("secondary")}>
        На главную
      </Link>
    </div>
  );
}
