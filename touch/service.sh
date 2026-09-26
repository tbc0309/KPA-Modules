#!/system/bin/sh
MODDIR=${0%/*}
while [ "$(getprop sys.boot_completed)" != 1 ]; do sleep 2; done
# Allow the controller service to finish its startup input-device reset.
sleep 10
# Retry startup failures only; a normal stop must remain stopped.
for ATTEMPT in 1 2 3; do
  "$MODDIR/bin/kpa_touch_guard" "$MODDIR" > "$MODDIR/startup.log" 2>&1
  [ "$?" = 0 ] && exit 0
  sleep 5
done
