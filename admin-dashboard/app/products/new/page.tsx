import Link from "next/link";
import { ArrowRight } from "lucide-react";
import DashboardShell from "@/components/layout/DashboardShell";
import ProductForm from "@/components/products/ProductForm";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export const dynamic = "force-dynamic";

export default function NewProductPage() {
  return (
    <DashboardShell title="title.new_product">
      <Link href="/products" className="mb-6 inline-flex items-center text-sm text-slate-500 hover:text-slate-900">
        <ArrowRight className="ml-1 h-4 w-4" /> رجوع للمنتجات
      </Link>
      <Card className="max-w-3xl">
        <CardHeader dir="rtl" className="text-right">
          <CardTitle>إنشاء منتج</CardTitle>
          <CardDescription>أضف كود اشتراك جديد للبيع في المتجر.</CardDescription>
        </CardHeader>
        <CardContent>
          <ProductForm />
        </CardContent>
      </Card>
    </DashboardShell>
  );
}
