#!/system/bin/sh

MODDIR=${0%/*}
NODE=/sys/sitronix_ts_attrs/stmt
LOG="$MODDIR/status.log"

# 每次启动保存简短记录，仅保留本次和上次日志。
# One short startup report per boot; retain only current and previous logs.
if [ -f "$LOG" ]; then
  mv -f "$LOG" "$MODDIR/status.previous.log"
fi
exec >>"$LOG" 2>&1
echo "$(date '+%F %T') waiting for Sitronix monitor control"

i=0
while [ "$i" -lt 60 ]; do
  if [ -w "$NODE" ]; then
    if echo 0 >"$NODE" && grep -q 'is_pause_mt = true' "$NODE"; then
      echo "$(date '+%F %T') Sitronix monitor paused"
      exit 0
    fi
    # 节点可能先于驱动初始化完成出现；在限定时间内重试，不常驻监控。
    # A node can appear before the driver is ready. Retry within the same
    # bounded startup window, then exit without a persistent watcher.
  fi
  i=$((i + 1))
  sleep 2
done

echo "$(date '+%F %T') compatible Sitronix control node not found"
exit 0
