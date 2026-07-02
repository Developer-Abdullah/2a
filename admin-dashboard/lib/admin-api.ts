import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

// Server-side client for the Admin API. The admin JWT lives only on the server (NextAuth
// session), so all data fetching happens in Server Components / Server Actions — never the
// browser. This also avoids CORS since the request originates server-to-server.
const ADMIN_BASE = process.env.BACKEND_API_URL || "http://localhost:8081";
const TENANT_SLUG = process.env.ADMIN_TENANT_SLUG || "store";

interface Envelope<T> {
  success: boolean;
  data?: T;
  error?: { message?: string };
}

async function authHeaders(): Promise<Record<string, string>> {
  const session = await getServerSession(authOptions);
  // Scope every request to the logged-in admin's own tenant. The platform owner has no tenant, so
  // fall back to the configured default (owners only use the non-tenant /platform endpoints).
  const slug = session?.user?.tenantSlug || TENANT_SLUG;
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${session?.accessToken ?? ""}`,
    "X-Tenant-Slug": slug,
  };
}

export async function adminGet<T>(path: string): Promise<T> {
  const res = await fetch(`${ADMIN_BASE}${path}`, {
    cache: "no-store",
    headers: await authHeaders(),
  });
  const body = (await res.json().catch(() => null)) as Envelope<T> | null;
  if (!res.ok || !body?.success) {
    throw new Error(body?.error?.message || `Request failed (${res.status})`);
  }
  return body.data as T;
}

export async function adminPost<T>(path: string, payload: unknown): Promise<T> {
  const res = await fetch(`${ADMIN_BASE}${path}`, {
    method: "POST",
    cache: "no-store",
    headers: await authHeaders(),
    body: JSON.stringify(payload),
  });
  const body = (await res.json().catch(() => null)) as Envelope<T> | null;
  if (!res.ok || !body?.success) {
    throw new Error(body?.error?.message || `Request failed (${res.status})`);
  }
  return body.data as T;
}

export async function adminPut<T>(path: string, payload: unknown): Promise<T> {
  const res = await fetch(`${ADMIN_BASE}${path}`, {
    method: "PUT",
    cache: "no-store",
    headers: await authHeaders(),
    body: JSON.stringify(payload),
  });
  const body = (await res.json().catch(() => null)) as Envelope<T> | null;
  if (!res.ok || !body?.success) {
    throw new Error(body?.error?.message || `Request failed (${res.status})`);
  }
  return body.data as T;
}
