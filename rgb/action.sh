#!/system/bin/sh

STATE=/data/adb/kpa_rgb_control.state
LOG=/data/adb/kpa_rgb_control.log
echo "KPA RGB Control"
echo "Current state:"
cat "$STATE" 2>/dev/null || echo "Not started"
echo
echo "Recent events:"
tail -n 12 "$LOG" 2>/dev/null || true
