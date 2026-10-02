#!/system/bin/sh

MODDIR=${0%/*}
NODE=/sys/sitronix_ts_attrs/stmt
LOG="$MODDIR/status.log"

exec >>"$LOG" 2>&1
echo "$(date '+%F %T') waiting for Sitronix monitor control"

i=0
while [ "$i" -lt 60 ]; do
  if [ -w "$NODE" ]; then
    if echo 0 >"$NODE" && grep -q 'is_pause_mt = true' "$NODE"; then
      echo "$(date '+%F %T') Sitronix monitor paused"
      exit 0
    fi
    echo "$(date '+%F %T') control node found but verification failed"
    exit 1
  fi
  i=$((i + 1))
  sleep 2
done

echo "$(date '+%F %T') compatible Sitronix control node not found"
exit 0
