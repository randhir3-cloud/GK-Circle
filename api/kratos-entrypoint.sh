#!/bin/sh
set -eu

if [ -n "${SECRETS_COOKIE:-}" ] && [ -z "${SECRETS_COOKIE_0:-}" ]; then
  export SECRETS_COOKIE_0="$SECRETS_COOKIE"
fi

if [ -n "${SECRETS_CIPHER:-}" ] && [ -z "${SECRETS_CIPHER_0:-}" ]; then
  export SECRETS_CIPHER_0="$SECRETS_CIPHER"
fi

unset SECRETS_COOKIE SECRETS_CIPHER

# Determine environment
APP_ENV_LOWER=$(echo "${APP_ENV:-development}" | tr '[:upper:]' '[:lower:]')
KRATOS_ENV_LOWER=$(echo "${KRATOS_ENV:-development}" | tr '[:upper:]' '[:lower:]')

if [ "$APP_ENV_LOWER" = "production" ] || [ "$KRATOS_ENV_LOWER" = "production" ] || [ "${RAILWAY_ENVIRONMENT:-}" = "production" ]; then
  echo "Starting in production mode: rendering Kratos config..."
  /usr/bin/kratos-config-renderer
else
  echo "Starting in development mode: using kratos.dev.yml..."
  cp /etc/config/kratos/kratos.dev.yml /tmp/kratos.yml
fi

exec /usr/bin/kratos-bin "$@"
