import { logoutAction } from "@/app/actions/auth";
import { Shell } from "@/components/shell";
import { roles } from "@/lib/labels";
import { getMe } from "@/lib/me";
import { adminNav, clinicNav } from "@/lib/nav";

export default async function AccountLayout({ children }: LayoutProps<"/account">) {
  const me = await getMe();
  const isPlatform = me.user.role === "superadmin";
  return (
    <Shell
      nav={isPlatform || !me.user.clinic_id ? adminNav() : clinicNav(me.user.clinic_id, me.user.role)}
      context={isPlatform ? { title: "Команда Jazyl", subtitle: "Все клиники" } : { title: me.clinic?.name ?? "Клиника" }}
      user={{ name: me.user.name, email: me.user.email, roleLabel: roles[me.user.role] }}
      logout={logoutAction}
    >
      {children}
    </Shell>
  );
}
