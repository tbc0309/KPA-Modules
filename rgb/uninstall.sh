#!/system/bin/sh

killall kpa_rgb_daemon 2>/dev/null
rm -f /data/adb/kpa_rgb_control.conf
rm -f /data/adb/kpa_rgb_control.state /data/adb/kpa_rgb_control.log
echo 0 > /sys/class/leds/red/brightness 2>/dev/null
echo 0 > /sys/class/leds/green/brightness 2>/dev/null
echo 0 > /sys/class/leds/blue/brightness 2>/dev/null
