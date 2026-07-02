"use server";

import { revalidatePath } from "next/cache";
import { adminPost, adminPut } from "@/lib/admin-api";

export interface ActionState {
  ok?: boolean;
  error?: string;
  message?: string;
  data?: Record<string, unknown>;
}

export async function createAppAction(_prev: ActionState, formData: FormData): Promise<ActionState> {
  try {
    const features = String(formData.get("features") || "")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const data = await adminPost<{ id: string }>("/v1/admin/apps", {
      bundle_identifier: String(formData.get("bundle_identifier") || "").trim(),
      name: String(formData.get("name") || "").trim(),
      category: String(formData.get("category") || "other"),
      description: String(formData.get("description") || ""),
      features,
      publish: formData.get("publish") === "on",
    });
    revalidatePath("/apps");
    return { ok: true, message: `App created (${data.id})`, data: { id: data.id } };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "Failed to create app" };
  }
}

export async function generateCodesAction(_prev: ActionState, formData: FormData): Promise<ActionState> {
  try {
    const data = await adminPost<{ codes: string[] }>("/v1/admin/activation/codes", {
      count: Number(formData.get("count") || 1),
      code_type: String(formData.get("code_type") || "usage_count"),
      device_type: String(formData.get("device_type") || "both"),
      max_devices: Number(formData.get("max_devices") || 1),
      max_uses: Number(formData.get("max_uses") || 1),
    });
    revalidatePath("/activation");
    return { ok: true, message: `Generated ${data.codes.length} code(s)`, data: { codes: data.codes } };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "Failed to generate codes" };
  }
}

export async function sendNotificationAction(_prev: ActionState, formData: FormData): Promise<ActionState> {
  try {
    await adminPost("/v1/admin/notifications", {
      title: String(formData.get("title") || "").trim(),
      body: String(formData.get("body") || "").trim(),
      type: String(formData.get("type") || "alert"),
    });
    revalidatePath("/notifications");
    return { ok: true, message: "Notification sent" };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "Failed to send notification" };
  }
}

// getUploadUrlAction returns a presigned PUT URL for the dashboard to upload a raw IPA to S3.
export async function getUploadUrlAction(appId: string): Promise<{ presigned_url: string; expected_key: string }> {
  return adminPost<{ presigned_url: string; expected_key: string }>(`/v1/admin/apps/${appId}/upload-url`, {});
}

export async function createVersionAction(_prev: ActionState, formData: FormData): Promise<ActionState> {
  try {
    const appId = String(formData.get("app_id") || "");
    await adminPost(`/v1/admin/apps/${appId}/versions`, {
      version: String(formData.get("version") || "").trim(),
      build_number: String(formData.get("build_number") || "1"),
      raw_ipa_s3_key: String(formData.get("raw_ipa_s3_key") || "").trim(),
      size_bytes: Number(formData.get("size_bytes") || 0),
    });
    revalidatePath(`/apps/${appId}`);
    return { ok: true, message: "Version queued for signing" };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "Failed to create version" };
  }
}

// --- Storefront catalog ---

export interface ProductPriceInput {
  currency: string;
  amount: number;
  compare_at: number | null;
}

export interface ProductFormInput {
  slug: string;
  name: string;
  subtitle: string;
  description: string;
  device_type: string;
  subscription_days: number;
  codes_per_unit: number;
  features: string[];
  terms: string[];
  video_url: string;
  is_published: boolean;
  sort_order: number;
  prices: ProductPriceInput[];
}

function actionError(e: unknown, fallback: string): ActionState {
  return { ok: false, error: e instanceof Error ? e.message : fallback };
}

export async function createProductAction(input: ProductFormInput): Promise<ActionState> {
  try {
    const data = await adminPost<{ id: string }>("/v1/admin/products", input);
    revalidatePath("/products");
    return { ok: true, message: "تم إنشاء المنتج", data: { id: data.id } };
  } catch (e) {
    return actionError(e, "تعذّر إنشاء المنتج");
  }
}

export async function updateProductAction(id: string, input: ProductFormInput): Promise<ActionState> {
  try {
    await adminPut(`/v1/admin/products/${id}`, input);
    revalidatePath("/products");
    revalidatePath(`/products/${id}`);
    return { ok: true, message: "تم حفظ التغييرات", data: { id } };
  } catch (e) {
    return actionError(e, "تعذّر حفظ المنتج");
  }
}

export async function setProductPublishedAction(id: string, published: boolean): Promise<ActionState> {
  try {
    await adminPost(`/v1/admin/products/${id}/publish`, { published });
    revalidatePath("/products");
    revalidatePath(`/products/${id}`);
    return { ok: true, message: published ? "تم النشر" : "تم إلغاء النشر" };
  } catch (e) {
    return actionError(e, "تعذّر تحديث حالة النشر");
  }
}

// confirmOrderAction manually marks an order paid+fulfilled (mints codes + queues the email). Used
// when the customer paid offline and the admin has verified it.
export async function confirmOrderAction(id: string): Promise<ActionState> {
  try {
    await adminPost(`/v1/admin/orders/${id}/confirm`, {});
    revalidatePath("/orders");
    revalidatePath(`/orders/${id}`);
    return { ok: true, message: "تم تأكيد الطلب وإصدار الأكواد" };
  } catch (e) {
    return actionError(e, "تعذّر تأكيد الطلب");
  }
}

// createTenantAction (platform owner) provisions a whole new merchant: tenant row, config, admin,
// certificate and a freshly migrated schema.
export async function createTenantAction(_prev: ActionState, formData: FormData): Promise<ActionState> {
  try {
    const slug = String(formData.get("slug") || "").trim().toLowerCase();
    await adminPost("/v1/admin/platform/tenants", {
      slug,
      name: String(formData.get("name") || "").trim(),
      admin_email: String(formData.get("admin_email") || "").trim(),
      admin_password: String(formData.get("admin_password") || ""),
    });
    revalidatePath("/owner");
    return { ok: true, message: `Merchant "${slug}" created`, data: { slug } };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "Failed to create merchant" };
  }
}
