FROM node:22-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ .
RUN pnpm run build

FROM caddy:2-alpine
COPY --from=frontend /app/dist /srv
COPY Caddyfile /etc/caddy/Caddyfile
