import { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

declare module "next-auth" {
  interface Session {
    accessToken: string;
    user: {
      id: string;
      email: string;
      role: string;
      tenantSlug: string;
      isOwner: boolean;
    };
  }

  interface User {
    id: string;
    email: string;
    role: string;
    accessToken: string;
    tenantSlug: string;
    isOwner: boolean;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    accessToken: string;
    id: string;
    role: string;
    tenantSlug: string;
    isOwner: boolean;
  }
}

export const authOptions: NextAuthOptions = {
  providers: [
    CredentialsProvider({
      name: "Admin Credentials",
      credentials: {
        email: { label: "Email", type: "email", placeholder: "admin@platform.com" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          throw new Error("Email and password are required.");
        }

        const apiUrl = process.env.BACKEND_API_URL || "http://localhost:8081";
        const res = await fetch(`${apiUrl}/v1/admin/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            email: credentials.email,
            password: credentials.password,
          }),
        });

        const json = await res.json();
        if (!res.ok || !json.success) {
          throw new Error(json.error?.message || "Invalid credentials provided.");
        }

        return {
          id: json.data.user.id,
          email: json.data.user.email,
          role: json.data.user.role,
          accessToken: json.data.access_token,
          tenantSlug: json.data.user.tenant_slug ?? "",
          isOwner: !!json.data.user.is_owner,
        };
      },
    }),
  ],
  session: {
    strategy: "jwt",
    maxAge: 60 * 60 * 24,
  },
  pages: {
    signIn: "/login",
  },
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id;
        token.role = user.role;
        token.accessToken = user.accessToken;
        token.tenantSlug = user.tenantSlug;
        token.isOwner = user.isOwner;
      }
      return token;
    },
    async session({ session, token }) {
      session.user.id = token.id;
      session.user.role = token.role;
      session.user.tenantSlug = token.tenantSlug;
      session.user.isOwner = token.isOwner;
      session.accessToken = token.accessToken;
      return session;
    },
  },
  secret: process.env.NEXTAUTH_SECRET,
};
