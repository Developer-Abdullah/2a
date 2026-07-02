import Link from "next/link";
import { ArrowRight } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import ProductForm from "@/components/products/ProductForm";
import PublishToggle from "@/components/products/PublishToggle";
import { adminGet } from "@/lib/admin-api";
import type { Product } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default async function EditProductPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let product: Product | null = null;
  let error = "";
  try {
    const data = await adminGet<{ product: Product }>(`/v1/admin/products/${id}`);
    product = data.product;
  } catch (e) {
    error = e instanceof Error ? e.message : "تعذّر تحميل المنتج";
  }

  return (
    <DashboardShell title="تعديل المنتج">
      <Link href="/products" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowRight className="ml-1 h-4 w-4" /> رجوع للمنتجات
      </Link>

      {error || !product ? (
        <p className="text-sm text-red-600">{error || "المنتج غير موجود"}</p>
      ) : (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader dir="rtl" className="text-right">
              <CardTitle>{product.name}</CardTitle>
              <CardDescription>عدّل بيانات المنتج وأسعاره.</CardDescription>
            </CardHeader>
            <CardContent>
              <ProductForm product={product} />
            </CardContent>
          </Card>

          <Card className="h-fit">
            <CardHeader dir="rtl" className="text-right">
              <CardTitle>حالة المنتج</CardTitle>
            </CardHeader>
            <CardContent dir="rtl" className="space-y-3 text-right text-sm text-slate-600">
              <div className="flex items-center justify-between">
                <span>عدد المبيعات</span>
                <span className="font-bold text-slate-900">{product.purchase_count.toLocaleString("ar-EG")}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>التقييم</span>
                <span className="font-bold text-slate-900">{product.rating_avg.toFixed(2)} ({product.rating_count})</span>
              </div>
              <PublishToggle id={product.id} published={product.is_published} />
            </CardContent>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
