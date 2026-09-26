#!/system/bin/sh
[ "$ARCH" = arm64 ] || abort 'ARM64 is required.'
set_perm_recursive "$MODPATH" 0 0 0755 0644
set_perm "$MODPATH/bin/kpa_touch_guard" 0 0 0755
set_perm "$MODPATH/service.sh" 0 0 0755
set_perm "$MODPATH/action.sh" 0 0 0755
set_perm "$MODPATH/uninstall.sh" 0 0 0755
ui_print 'Hold MODE for 2 seconds to toggle touch. Reboot always restores touch.'
