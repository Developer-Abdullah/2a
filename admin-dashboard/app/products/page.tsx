import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { Product } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";

export const dynamic = "force-dynamic";

function priceLabel(p: Product): string {
  if (p.prices.length === 0) return "—";
  return p.prices.map((pr) => `${pr.amount} ${pr.currency}`).join(" · ");
}

export default async function ProductsPage() {
  let products: Product[] = [];
  let error = "";
  try {
    const data = await adminGet<{ items: Product[] }>("/v1/admin/products");
    products = data.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : "تعذّر تحميل المنتجات";
  }

  return (
    <DashboardShell title="المنتجات">
      <div className="mb-6 flex items-center justify-between">
        <p className="text-sm text-slate-500">{products.length} منتج</p>
        <Link href="/products/new">
          <Button>منتج جديد</Button>
        </Link>
      </div>

      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : products.length === 0 ? (
        <p className="text-sm text-slate-500">لا توجد منتجات بعد. أنشئ أول منتج.</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>الاسم</TableHead>
                  <TableHead>المُعرّف</TableHead>
                  <TableHead>الأسعار</TableHead>
                  <TableHead>المبيعات</TableHead>
                  <TableHead>الحالة</TableHead>
                  <TableHead className="text-right">إجراء</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {products.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-medium text-slate-900">{p.name}</TableCell>
                    <TableCell className="text-slate-500">{p.slug}</TableCell>
                    <TableCell className="text-slate-600">{priceLabel(p)}</TableCell>
                    <TableCell className="text-slate-600">{p.purchase_count.toLocaleString("ar-EG")}</TableCell>
                    <TableCell>
                      {p.is_published ? <Badge variant="success">منشور</Badge> : <Badge variant="muted">مسودة</Badge>}
                    </TableCell>
                    <TableCell className="text-right">
                      <Link href={`/products/${p.id}`} className="text-sm font-medium text-blue-600 hover:underline">
                        تعديل
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </DashboardShell>
  );
}
