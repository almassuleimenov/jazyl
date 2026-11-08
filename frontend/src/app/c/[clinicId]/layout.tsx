import { notFound } from "next/navigation";

import { logoutAction } from "@/app/actions/auth";
import { Shell } from "@/components/shell";
import { ApiError, api } from "@/lib/api";
import { roles } from "@/lib/labels";
import { getMe } from "@/lib/me";
import { clinicNav } from "@/lib/nav";
import type { Lead, Page } from "@/lib/types";
import { getClinic } from "./data";

export default async function ClinicLayout({ children, params }: LayoutProps<"/c/[clinicId]">) {
  const { clinicId } = await params;
  const me = await getMe();
  let clinic;
  try {
    clinic = await getClinic(clinicId);
  } catch (e) {
    // The API answers 404 for clinics outside the caller's tenant.
    if (e instanceof ApiError && (e.status === 404 || e.status === 422)) notFound();
    throw e;
  }
  // The badge is a nicety: an API error here must not break the whole clinic area.
  const leads = await api<Page<Lead>>(`/clinics/${encodeURIComponent(clinicId)}/leads?status=new&limit=50`).catch((e) => {
    if (e instanceof ApiError) return { items: [] as Lead[] };
    throw e; // redirects (session expired) must propagate
  });

  return (
    <Shell
      nav={clinicNav(clinicId, me.user.role, leads.items.length)}
      context={{ title: clinic.name, subtitle: [clinic.city, clinic.address].filter(Boolean).join(", ") || undefined }}
      back={me.user.role === "superadmin" ? { href: "/admin", label: "Все клиники" } : undefined}
      user={{ name: me.user.name, email: me.user.email, roleLabel: roles[me.user.role] }}
      logout={logoutAction}
    >
      {children}
    </Shell>
  );
}
