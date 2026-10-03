#!/system/bin/sh
MODDIR=${0%/*}
while [ "$(getprop sys.boot_completed)" != 1 ]; do sleep 2; done
# 等待手柄服务完成启动时的输入设备重置。
# Allow the controller service to finish its startup input-device reset.
sleep 10
# 仅重试启动失败；正常停止后不自动重启。
# Retry startup failures only; a normal stop must remain stopped.
for ATTEMPT in 1 2 3; do
  "$MODDIR/bin/kpa_touch_guard" "$MODDIR" > "$MODDIR/startup.log" 2>&1
  [ "$?" = 0 ] && exit 0
  sleep 5
done
