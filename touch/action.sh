#!/system/bin/sh
MODDIR=${0%/*}
PID=$(cat "$MODDIR/guard.pid" 2>/dev/null)
case "$PID" in ''|*[!0-9]*) exit 0;; esac
if [ "$(readlink /proc/$PID/exe)" = "$MODDIR/bin/kpa_touch_guard" ]; then
  kill -TERM "$PID"
  echo 'Touch restored. Restart the module service or reboot to resume MODE monitoring.'
fi
