#!/system/bin/sh

set_perm "$MODPATH/service.sh" 0 0 0755
set_perm "$MODPATH/action.sh" 0 0 0755
set_perm "$MODPATH/uninstall.sh" 0 0 0755

ui_print 'KPA Sitronix Fix'
ui_print 'Pauses only the driver health-monitor loop; touch input remains enabled.'
