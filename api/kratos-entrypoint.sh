#!/bin/sh
set -eu

if [ -n "${SECRETS_COOKIE:-}" ] && [ -z "${SECRETS_COOKIE_0:-}" ]; then
  export SECRETS_COOKIE_0="$SECRETS_COOKIE"
fi

if [ -n "${SECRETS_CIPHER:-}" ] && [ -z "${SECRETS_CIPHER_0:-}" ]; then
  export SECRETS_CIPHER_0="$SECRETS_CIPHER"
fi

unset SECRETS_COOKIE SECRETS_CIPHER

exec /usr/bin/kratos-bin "$@"
