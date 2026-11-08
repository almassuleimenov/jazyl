import { redirect } from "next/navigation";

import { getMe } from "@/lib/me";

// Sends each role to its home: Jazyl team → all clinics, clinic staff → their clinic.
export default async function Home() {
  const me = await getMe();
  if (me.user.role === "superadmin") redirect("/admin");
  if (me.user.clinic_id) redirect(`/c/${me.user.clinic_id}`);
  redirect("/account");
}
