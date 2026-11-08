import type { NavItem } from "@/components/shell";
import type { Role } from "./types";

export function adminNav(): NavItem[] {
  return [{ href: "/admin", label: "Клиники", icon: "clinics", exact: true }];
}

export function clinicNav(clinicId: string, role: Role, newLeads = 0): NavItem[] {
  const base = `/c/${clinicId}`;
  const items: NavItem[] = [
    { href: base, label: "Обзор", icon: "overview", exact: true },
    { href: `${base}/appointments`, label: "Записи", icon: "appointments" },
    { href: `${base}/leads`, label: "Заявки", icon: "leads", badge: newLeads || undefined },
    { href: `${base}/chats`, label: "Переписка", icon: "chats" },
  ];
  if (role === "owner" || role === "superadmin") {
    items.push(
      { href: `${base}/team`, label: "Команда и доступ", icon: "team" },
      { href: `${base}/site`, label: "Сайт", icon: "site" },
      { href: `${base}/settings`, label: "Настройки", icon: "settings" },
    );
  }
  return items;
}
