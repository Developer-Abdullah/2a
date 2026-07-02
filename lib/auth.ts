import { NextAuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

// Strictly type the NextAuth session to include our Go API JWT and custom user fields
declare module "next-auth" {
  interface Session {
    accessToken: string;
    user: {
      id: string;
      email: string;
      role: string;
    };
  }

  interface User {
    id: string;
    email: string;
    role: string;
    accessToken: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    accessToken: string;
    id: string;
    role: string;
  }
}

export const authOptions: NextAuthOptions = {
  providers: [
    CredentialsProvider({
      name: "Admin Credentials",
      credentials: {
        email: { label: "Email", type: "email", placeholder: "admin@store.com" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          throw new Error("Email and password are required.");
        }

        // BACKEND_API_URL should be an internal Docker network URL in production 
        // (e.g., http://admin-api:8081) to keep auth traffic off the public internet.
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
          // Throwing an error here triggers the NextAuth error flow, displaying it on the frontend
          throw new Error(json.error?.message || "Invalid credentials provided.");
        }

        // Return the mapped User object containing the Go API access token
        return {
          id: json.data.user.id,
          email: json.data.user.email,
          role: json.data.user.role,
          accessToken: json.data.access_token,
        };
      },
    }),
  ],
  session: {
    strategy: "jwt",
    maxAge: 60 * 60 * 24, // 1 Day (Should match or coordinate with Go API refresh token limits)
  },
  pages: {
    signIn: "/login",
  },
  callbacks: {
    // Triggered whenever a JWT is created or updated
    async jwt({ token, user }) {
      // Initial sign-in: Inject the Go API JWT into the NextAuth encrypted token
      if (user) {
        token.id = user.id;
        token.role = user.role;
        token.accessToken = user.accessToken;
      }
      return token;
    },
    // Triggered whenever the session is checked by the client or server components
    async session({ session, token }) {
      if (token) {
        session.user.id = token.id;
        session.user.role = token.role;
        session.accessToken = token.accessToken;
      }
      return session;
    },
  },
  secret: process.env.NEXTAUTH_SECRET,
};