import { withAuth } from "next-auth/middleware";
import type { NextFetchEvent } from "next/server";
import type { NextRequestWithAuth } from "next-auth/middleware";

const authProxy = withAuth({
  callbacks: {
    authorized: ({ token }) => !!token,
  },
});

export function proxy(request: NextRequestWithAuth, event: NextFetchEvent) {
  return authProxy(request, event);
}

export const config = {
  matcher: ["/dashboard/:path*", "/products/:path*", "/orders/:path*"],
};
