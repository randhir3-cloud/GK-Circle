FROM nginx:1.27-alpine

COPY nginx.railway.conf /etc/nginx/nginx.conf

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

CMD ["sh", "-c", "export RESOLVER=$(awk '/nameserver/ {print $2; exit}' /etc/resolv.conf) && sed -i \"s/127.0.0.11/$RESOLVER/g\" /etc/nginx/nginx.conf && exec nginx -g 'daemon off;'"]

