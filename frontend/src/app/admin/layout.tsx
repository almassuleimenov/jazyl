import { notFound } from "next/navigation";

import { logoutAction } from "@/app/actions/auth";
import { Shell } from "@/components/shell";
import { roles } from "@/lib/labels";
import { getMe } from "@/lib/me";
import { adminNav } from "@/lib/nav";

export default async function AdminLayout({ children }: LayoutProps<"/admin">) {
  const me = await getMe();
  // The API enforces this too (403); hiding the platform area from clinic staff is UX only.
  if (me.user.role !== "superadmin") notFound();
  return (
    <Shell
      nav={adminNav()}
      context={{ title: "Команда Jazyl", subtitle: "Все клиники" }}
      user={{ name: me.user.name, email: me.user.email, roleLabel: roles[me.user.role] }}
      logout={logoutAction}
    >
      {children}
    </Shell>
  );
}
