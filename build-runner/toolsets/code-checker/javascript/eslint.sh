#!/bin/sh
js_root=/opt/unifi-stubd-code-checker-javascript
exec "$js_root/node_modules/.bin/eslint" \
  --config "$js_root/eslint.config.mjs" "$@"
