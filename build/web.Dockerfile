FROM node:22-alpine AS deps

WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm install

FROM node:22-alpine AS builder
WORKDIR /app

# NEXT_PUBLIC_* sont résolus au build — doit être passé en ARG
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL

COPY --from=deps /app/node_modules ./node_modules
COPY web ./
RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
