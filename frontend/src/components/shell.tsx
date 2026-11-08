"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";
import {
  ArrowLeft,
  Building2,
  CalendarDays,
  Globe,
  Inbox,
  LayoutDashboard,
  LogOut,
  Menu,
  MessagesSquare,
  Settings,
  ShieldCheck,
  Users,
  X,
} from "lucide-react";

import { Logo } from "./logo";
import { cx } from "./ui";

const icons = {
  overview: LayoutDashboard,
  appointments: CalendarDays,
  leads: Inbox,
  chats: MessagesSquare,
  team: Users,
  settings: Settings,
  site: Globe,
  clinics: Building2,
  account: ShieldCheck,
};

export type NavItem = { href: string; label: string; icon: keyof typeof icons; badge?: number; exact?: boolean };

type Props = {
  nav: NavItem[];
  context: { title: string; subtitle?: string };
  back?: { href: string; label: string };
  user: { name: string; email: string; roleLabel: string };
  logout: () => Promise<void>;
  children: ReactNode;
};

function isActive(pathname: string, item: NavItem) {
  return item.exact ? pathname === item.href : pathname === item.href || pathname.startsWith(item.href + "/");
}

function Sidebar({ nav, context, back, user, logout, onNavigate }: Omit<Props, "children"> & { onNavigate?: () => void }) {
  const pathname = usePathname();
  return (
    <div className="flex h-full flex-col bg-brand text-brand-ink">
      <div className="flex h-16 shrink-0 items-center px-5">
        <Link href="/" onClick={onNavigate} className="rounded-sm focus-visible:outline-action">
          <Logo />
        </Link>
      </div>

      <div className="px-3">
        {back ? (
          <Link
            href={back.href}
            onClick={onNavigate}
            className="mb-2 flex items-center gap-2 rounded-control px-2 py-1.5 text-[13px] text-brand-ink/90 hover:bg-white/8 hover:text-white focus-visible:outline-action"
          >
            <ArrowLeft className="size-3.5" aria-hidden />
            {back.label}
          </Link>
        ) : null}
        <div className="rounded-control border border-white/12 bg-white/6 px-3 py-2.5">
          <p className="truncate text-sm font-semibold text-white">{context.title}</p>
          {context.subtitle ? <p className="mt-0.5 truncate text-xs text-brand-ink/85">{context.subtitle}</p> : null}
        </div>
      </div>

      <nav aria-label="Разделы" className="mt-4 flex-1 overflow-y-auto px-3">
        <ul className="flex flex-col gap-0.5">
          {nav.map((item) => {
            const Icon = icons[item.icon];
            const active = isActive(pathname, item);
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  onClick={onNavigate}
                  aria-current={active ? "page" : undefined}
                  className={cx(
                    "group flex h-10 items-center gap-3 rounded-control px-3 text-sm transition-colors duration-150 focus-visible:outline-action",
                    active ? "bg-white text-brand font-semibold shadow-panel" : "text-brand-ink hover:bg-white/8 hover:text-white",
                  )}
                >
                  <Icon className={cx("size-[18px] shrink-0", active ? "text-brand" : "text-brand-ink/80 group-hover:text-white")} aria-hidden />
                  <span className="flex-1 truncate">{item.label}</span>
                  {item.badge ? (
                    <span className="num rounded-full bg-action px-1.5 py-px text-[11px] font-semibold text-ink">{item.badge}</span>
                  ) : null}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="border-t border-white/10 p-3">
        <Link
          href="/account"
          onClick={onNavigate}
          aria-current={pathname === "/account" ? "page" : undefined}
          aria-label={`Аккаунт и безопасность: ${user.name || user.email}`}
          className={cx(
            "flex items-center gap-3 rounded-control px-2 py-2 transition-colors hover:bg-white/8 focus-visible:outline-action",
            pathname === "/account" && "bg-white/10",
          )}
        >
          <span className="grid size-8 shrink-0 place-items-center rounded-full bg-white/12 text-[13px] font-semibold text-white" aria-hidden>
            {(user.name || user.email).slice(0, 1).toUpperCase()}
          </span>
          <span className="min-w-0 flex-1">
            <span className="block truncate text-sm text-white">{user.name || user.email}</span>
            <span className="block truncate text-xs text-brand-ink/85">{user.roleLabel}</span>
          </span>
        </Link>
        <form action={logout}>
          <button
            type="submit"
            className="mt-1 flex h-9 w-full items-center gap-3 rounded-control px-3 text-sm text-brand-ink hover:bg-white/8 hover:text-white focus-visible:outline-action"
          >
            <LogOut className="size-4" aria-hidden />
            Выйти
          </button>
        </form>
      </div>
    </div>
  );
}

export function Shell(props: Props) {
  // The drawer closes through each link's onNavigate, Escape, or the backdrop.
  const [open, setOpen] = useState(false);
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open]);

  return (
    <div className="min-h-dvh overflow-x-clip lg:pl-60">
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-60 lg:block">
        <Sidebar {...props} />
      </aside>

      {/* Phone: top bar + drawer */}
      <div className="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-brand-strong bg-brand px-4 lg:hidden">
        <Link href="/" aria-label="На главную">
          <Logo className="text-[20px]" />
        </Link>
        <p className="mx-3 min-w-0 flex-1 truncate text-right text-sm text-brand-ink">{props.context.title}</p>
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="grid size-10 place-items-center rounded-control text-white hover:bg-white/10 focus-visible:outline-action"
          aria-label="Открыть меню"
          aria-expanded={open}
        >
          <Menu className="size-5" aria-hidden />
        </button>
      </div>
      <div
        className={cx("fixed inset-0 z-40 lg:hidden", open ? "pointer-events-auto" : "pointer-events-none")}
        aria-hidden={!open}
        inert={!open}
      >
        <div
          className={cx("absolute inset-0 bg-ink/40 transition-opacity duration-200", open ? "opacity-100" : "opacity-0")}
          onClick={() => setOpen(false)}
        />
        <div
          role="dialog"
          aria-modal="true"
          aria-label="Меню"
          className={cx(
            "absolute inset-y-0 left-0 w-[min(18rem,86vw)] shadow-pop transition-transform duration-300 ease-out",
            open ? "translate-x-0" : "-translate-x-full",
          )}
        >
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="absolute top-3 right-3 z-10 grid size-10 place-items-center rounded-control text-white hover:bg-white/10 focus-visible:outline-action"
            aria-label="Закрыть меню"
          >
            <X className="size-5" aria-hidden />
          </button>
          <Sidebar {...props} onNavigate={() => setOpen(false)} />
        </div>
      </div>

      <main className="mx-auto w-full min-w-0 max-w-[1240px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">{props.children}</main>
    </div>
  );
}
