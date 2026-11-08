import "server-only";

import { cache } from "react";

import { api } from "./api";
import type { Me } from "./types";

/** Current user, memoised per request (layouts and pages share one /auth/me call). */
export const getMe = cache(() => api<Me>("/auth/me"));
