FROM node:20-alpine AS deps
WORKDIR /app
COPY storefront/package.json storefront/package-lock.json* ./
# npm install (not ci) so platform-specific optional deps resolve even when the lock file was
# generated on a different OS.
RUN npm install --no-audit --no-fund

FROM node:20-alpine AS builder
WORKDIR /app
# NEXT_PUBLIC_* are inlined at build time, so contact links must be present during `npm run build`.
ARG NEXT_PUBLIC_WHATSAPP=""
ARG NEXT_PUBLIC_TELEGRAM=""
ENV NEXT_PUBLIC_WHATSAPP=$NEXT_PUBLIC_WHATSAPP
ENV NEXT_PUBLIC_TELEGRAM=$NEXT_PUBLIC_TELEGRAM
COPY storefront/ ./
COPY --from=deps /app/node_modules ./node_modules
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
ENV PORT 3001
ENV HOSTNAME 0.0.0.0
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3001
CMD ["node", "server.js"]
