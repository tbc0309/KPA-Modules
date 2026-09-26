#!/system/bin/sh

ui_print "- KONKR Pocket Advance MYuppy Font"
ui_print "- Systemless installation; stock files remain unchanged"

# Inspect both installed modules and staged updates before applying overlays.
. "$MODPATH/check-conflicts.sh"
check_font_conflicts /data/adb/modules /data/adb/modules_update

# Abort early if the release archive is incomplete.
[ -f "$MODPATH/system/etc/fonts.xml" ] || abort "Missing system/etc/fonts.xml"
[ -d "$MODPATH/system/fonts" ] || abort "Missing system/fonts"

# Directories must be executable and font/config files must be readable.
set_perm_recursive "$MODPATH/system" 0 0 0755 0644
