import Link from "next/link";

export default function NotFound() {
  return (
    <div className="container-page py-24 text-center">
      <h1 className="text-5xl font-extrabold text-brand-700">٤٠٤</h1>
      <p className="mt-4 text-lg text-slate-500">الصفحة غير موجودة.</p>
      <Link href="/" className="btn-primary mt-8">العودة للرئيسية</Link>
    </div>
  );
}
