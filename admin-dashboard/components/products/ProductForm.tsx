"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  ProductFormInput,
  createProductAction,
  updateProductAction,
  type ActionState,
} from "@/lib/actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Product } from "@/lib/types";

const CURRENCIES = ["EGP", "KWD"] as const;

function priceOf(product: Product | undefined, currency: string, field: "amount" | "compare_at"): string {
  const p = product?.prices.find((x) => x.currency === currency);
  if (!p) return "";
  const v = field === "amount" ? p.amount : p.compare_at;
  return v == null ? "" : String(v);
}

// ProductForm creates a new product or edits an existing one (when `product` is provided). Prices are
// captured per supported currency (EGP, KWD); features and terms are one-per-line textareas.
export default function ProductForm({ product }: { product?: Product }) {
  const router = useRouter();
  const isEdit = !!product;
  const [state, setState] = useState<ActionState>({});
  const [pending, setPending] = useState(false);
  // image_s3_key round-trips through the form; product.image_url IS the stored key on the admin side.
  const [imageKey, setImageKey] = useState(product?.image_url || "");
  const [imgPreview, setImgPreview] = useState("");
  const [uploading, setUploading] = useState(false);
  const [uploadErr, setUploadErr] = useState("");

  async function onImageChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !product) return;
    setUploadErr("");
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`/api/product-image/${product.id}`, { method: "POST", body: fd });
      const data = await res.json();
      if (!res.ok || !data.key) throw new Error("upload failed");
      setImageKey(data.key);
      setImgPreview(URL.createObjectURL(file));
    } catch {
      setUploadErr("تعذّر رفع الصورة. حاول مرة أخرى.");
    } finally {
      setUploading(false);
    }
  }

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const num = (k: string, d = 0) => {
      const n = Number(fd.get(k));
      return Number.isFinite(n) && n > 0 ? n : d;
    };
    const lines = (k: string) =>
      String(fd.get(k) || "")
        .split("\n")
        .map((s) => s.trim())
        .filter(Boolean);

    const prices = CURRENCIES.map((cur) => {
      const amount = Number(fd.get(`amount_${cur}`));
      const compareRaw = fd.get(`compare_${cur}`);
      const compare = compareRaw ? Number(compareRaw) : NaN;
      return {
        currency: cur,
        amount: Number.isFinite(amount) ? amount : 0,
        compare_at: Number.isFinite(compare) && compare > 0 ? compare : null,
      };
    }).filter((p) => p.amount > 0);

    const input: ProductFormInput = {
      slug: String(fd.get("slug") || "").trim(),
      name: String(fd.get("name") || "").trim(),
      subtitle: String(fd.get("subtitle") || "").trim(),
      description: String(fd.get("description") || "").trim(),
      device_type: String(fd.get("device_type") || "both"),
      subscription_days: num("subscription_days", 365),
      codes_per_unit: num("codes_per_unit", 1),
      features: lines("features"),
      terms: lines("terms"),
      video_url: String(fd.get("video_url") || "").trim(),
      image_s3_key: imageKey,
      is_published: fd.get("is_published") === "on",
      sort_order: Number(fd.get("sort_order") || 0),
      prices,
    };

    setPending(true);
    const res = isEdit ? await updateProductAction(product!.id, input) : await createProductAction(input);
    setPending(false);
    setState(res);
    if (res.ok) {
      const id = (res.data?.id as string) || product?.id;
      router.push("/products");
      router.refresh();
      void id;
    }
  }

  const label = "text-sm font-medium text-slate-700";
  const field = "flex w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm";

  return (
    <form onSubmit={onSubmit} dir="rtl" className="space-y-5 text-right">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1">
          <label className={label}>الاسم</label>
          <Input name="name" required defaultValue={product?.name} placeholder="اشتراك تطبيقات بلس" />
        </div>
        <div className="space-y-1">
          <label className={label}>المُعرّف (slug)</label>
          <Input name="slug" required defaultValue={product?.slug} placeholder="PdDWGXW" />
        </div>
      </div>

      <div className="space-y-1">
        <label className={label}>العنوان الفرعي</label>
        <Input name="subtitle" defaultValue={product?.subtitle} placeholder="كود تفعيل فوري" />
      </div>

      <div className="space-y-1">
        <label className={label}>الوصف</label>
        <textarea name="description" rows={3} className={field} defaultValue={product?.description} placeholder="وصف المنتج" />
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <div className="space-y-1">
          <label className={label}>الجهاز</label>
          <select name="device_type" defaultValue={product?.device_type || "both"} className={field}>
            <option value="both">آيفون / آيباد</option>
            <option value="iphone">آيفون</option>
            <option value="ipad">آيباد</option>
          </select>
        </div>
        <div className="space-y-1">
          <label className={label}>مدة الاشتراك (أيام)</label>
          <Input name="subscription_days" type="number" min={1} defaultValue={product?.subscription_days ?? 365} />
        </div>
        <div className="space-y-1">
          <label className={label}>أكواد لكل وحدة</label>
          <Input name="codes_per_unit" type="number" min={1} defaultValue={product?.codes_per_unit ?? 1} />
        </div>
      </div>

      {/* Prices per currency */}
      <div className="rounded-lg border border-slate-200 p-4">
        <h3 className="mb-3 text-sm font-semibold text-slate-900">الأسعار</h3>
        <div className="space-y-3">
          {CURRENCIES.map((cur) => (
            <div key={cur} className="grid grid-cols-[auto,1fr,1fr] items-center gap-3">
              <span className="w-12 text-sm font-bold text-slate-600">{cur}</span>
              <div className="space-y-1">
                <label className="text-xs text-slate-500">السعر</label>
                <Input name={`amount_${cur}`} type="number" step="0.001" min={0} defaultValue={priceOf(product, cur, "amount")} placeholder="0.000" />
              </div>
              <div className="space-y-1">
                <label className="text-xs text-slate-500">قبل الخصم (اختياري)</label>
                <Input name={`compare_${cur}`} type="number" step="0.001" min={0} defaultValue={priceOf(product, cur, "compare_at")} placeholder="—" />
              </div>
            </div>
          ))}
        </div>
        <p className="mt-2 text-xs text-slate-400">اترك السعر فارغًا لعدم عرض المنتج بتلك العملة.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1">
          <label className={label}>المميزات (سطر لكل ميزة)</label>
          <textarea name="features" rows={5} className={field} defaultValue={product?.features.join("\n")} placeholder={"تفعيل فوري\nبدون جيلبريك"} />
        </div>
        <div className="space-y-1">
          <label className={label}>الشروط (سطر لكل شرط)</label>
          <textarea name="terms" rows={5} className={field} defaultValue={product?.terms.join("\n")} placeholder={"غير قابل للاسترجاع"} />
        </div>
      </div>

      {/* Product image */}
      <div className="rounded-lg border border-slate-200 p-4">
        <h3 className="mb-2 text-sm font-semibold text-slate-900">صورة المنتج</h3>
        {isEdit ? (
          <div className="flex items-center gap-4">
            <div className="flex h-20 w-20 flex-none items-center justify-center overflow-hidden rounded-lg bg-slate-100 text-center ring-1 ring-slate-200">
              {imgPreview ? (
                <img src={imgPreview} alt="" className="h-full w-full object-cover" />
              ) : imageKey ? (
                <span className="px-1 text-[10px] font-medium text-emerald-600">صورة مرفوعة</span>
              ) : (
                <span className="text-xs text-slate-400">لا صورة</span>
              )}
            </div>
            <div>
              <input type="file" accept="image/*" onChange={onImageChange} disabled={uploading} className="text-sm" />
              <p className="mt-1 text-xs text-slate-400">{uploading ? "جارٍ الرفع…" : imageKey ? "تم رفع صورة ✓ — احفظ لتطبيقها" : "PNG/JPG — تُرفع مباشرة للتخزين"}</p>
              {uploadErr ? <p className="text-xs text-red-600">{uploadErr}</p> : null}
            </div>
          </div>
        ) : (
          <p className="text-xs text-slate-400">احفظ المنتج أولًا ثم افتحه للتعديل لرفع صورة.</p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1">
          <label className={label}>رابط فيديو الشرح (اختياري)</label>
          <Input name="video_url" defaultValue={product?.video_url} placeholder="https://…" />
        </div>
        <div className="space-y-1">
          <label className={label}>ترتيب العرض</label>
          <Input name="sort_order" type="number" defaultValue={product?.sort_order ?? 0} />
        </div>
      </div>

      <label className="flex items-center gap-2 text-sm text-slate-700">
        <input type="checkbox" name="is_published" defaultChecked={product ? product.is_published : true} /> نشر في المتجر
      </label>

      {state.error ? <p className="text-sm text-red-600">{state.error}</p> : null}
      {state.message ? <p className="text-sm text-green-600">{state.message}</p> : null}

      <div className="flex gap-3">
        <Button type="submit" disabled={pending}>
          {pending ? "جارٍ الحفظ…" : isEdit ? "حفظ التغييرات" : "إنشاء المنتج"}
        </Button>
        <Button type="button" variant="outline" onClick={() => router.push("/products")}>
          إلغاء
        </Button>
      </div>
    </form>
  );
}
