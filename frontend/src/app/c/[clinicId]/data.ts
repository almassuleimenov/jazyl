import "server-only";

import { cache } from "react";

import { api } from "@/lib/api";
import type { Clinic } from "@/lib/types";

export const getClinic = cache((clinicId: string) => api<Clinic>(`/clinics/${encodeURIComponent(clinicId)}/`));
