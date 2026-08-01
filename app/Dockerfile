FROM node:22-slim AS build

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:22-alpine AS runtime

WORKDIR /app
RUN apk add --no-cache nginx supervisor wget
RUN mkdir -p /run/nginx

COPY --from=build /app/.output ./.output
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/supervisiord.conf /etc/supervisord.conf

ENV HOST=0.0.0.0
ENV PORT=4000
ENV NODE_ENV=production

EXPOSE 5000
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -qO- http://localhost:5000/favicon.ico >/dev/null || exit 1

CMD ["supervisord", "-c", "/etc/supervisord.conf"]
