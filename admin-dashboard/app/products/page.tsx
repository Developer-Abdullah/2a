import Link from "next/link";
import DashboardShell from "@/components/layout/DashboardShell";
import { adminGet } from "@/lib/admin-api";
import type { Product } from "@/lib/types";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { getT } from "@/lib/locale-server";

export const dynamic = "force-dynamic";

function priceLabel(p: Product): string {
  if (p.prices.length === 0) return "—";
  return p.prices.map((pr) => `${pr.amount} ${pr.currency}`).join(" · ");
}

export default async function ProductsPage() {
  const t = await getT();
  let products: Product[] = [];
  let error = "";
  try {
    const data = await adminGet<{ items: Product[] }>("/v1/admin/products");
    products = data.items ?? [];
  } catch (e) {
    error = e instanceof Error ? e.message : t("dash.load_error");
  }

  return (
    <DashboardShell title="products.title">
      <div className="mb-6 flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{products.length} {t("products.count")}</p>
        <Link href="/products/new">
          <Button>{t("products.new")}</Button>
        </Link>
      </div>

      {error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : products.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("products.count")}: 0</p>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("products.col.name")}</TableHead>
                  <TableHead>{t("products.col.id")}</TableHead>
                  <TableHead>{t("products.col.prices")}</TableHead>
                  <TableHead>{t("products.col.sales")}</TableHead>
                  <TableHead>{t("products.col.status")}</TableHead>
                  <TableHead className="text-end"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {products.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-medium text-foreground">{p.name}</TableCell>
                    <TableCell className="text-muted-foreground">{p.slug}</TableCell>
                    <TableCell className="text-foreground/80">{priceLabel(p)}</TableCell>
                    <TableCell className="text-foreground/80">{p.purchase_count.toLocaleString()}</TableCell>
                    <TableCell>
                      {p.is_published ? (
                        <Badge variant="success">{t("products.published")}</Badge>
                      ) : (
                        <Badge variant="muted">{t("products.draft")}</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-end">
                      <Link href={`/products/${p.id}`} className="text-sm font-medium text-brand-700 hover:underline dark:text-brand-300">
                        {t("common.edit")}
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
