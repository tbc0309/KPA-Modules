#!/system/bin/sh

NODE=/sys/sitronix_ts_attrs/stmt

if [ -w "$NODE" ]; then
  echo 1 >"$NODE"
fi
