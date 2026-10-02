#!/system/bin/sh

NODE=/sys/sitronix_ts_attrs/stmt

if [ ! -w "$NODE" ]; then
  echo 'Compatible Sitronix control node was not found.'
  exit 1
fi

echo 0 >"$NODE"
cat "$NODE"
