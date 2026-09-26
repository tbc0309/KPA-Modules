#!/system/bin/sh

ui_print "========================================"
ui_print " KPA RGB Control v1.0.0"
ui_print " KONKR Pocket Advance RGB performance LED"
ui_print "========================================"

CONF=/data/adb/kpa_rgb_control.conf
cp "$MODPATH/config.default.conf" "$CONF"
chmod 0644 "$CONF"
rm -f /data/adb/kpa_rgb_control.state /data/adb/kpa_rgb_control.log
ui_print "- Created config: $CONF"

set_perm "$MODPATH/service.sh" 0 0 0755
set_perm "$MODPATH/bin/kpa_rgb_daemon" 0 0 0755
set_perm "$MODPATH/action.sh" 0 0 0755
set_perm "$MODPATH/uninstall.sh" 0 0 0755
set_perm "$MODPATH/config.default.conf" 0 0 0644
set_perm "$MODPATH/README_CN_EN.txt" 0 0 0644
ui_print "- Reboot to enable the RGB controller."
