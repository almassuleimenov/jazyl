import { Logo } from "@/components/logo";
import { logoutAction } from "@/app/actions/auth";

// Minimal frame for the steps a session must finish before the cabinet opens.
export default function SetupLayout({ children }: LayoutProps<"/setup">) {
  return (
    <div className="flex min-h-dvh flex-col">
      <header className="flex h-16 items-center justify-between border-b border-line bg-surface px-4 sm:px-6">
        <Logo tone="dark" />
        <form action={logoutAction}>
          <button type="submit" className="rounded-control px-3 py-2 text-sm text-ink-2 hover:bg-sunk hover:text-ink">
            Выйти
          </button>
        </form>
      </header>
      <main className="mx-auto w-full max-w-[640px] flex-1 px-4 py-10 sm:py-14">{children}</main>
    </div>
  );
}
