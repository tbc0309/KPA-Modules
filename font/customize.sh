#!/system/bin/sh

ui_print "- KONKR Pocket Advance MYuppy Font"
ui_print "- Systemless installation; stock files remain unchanged"

# 应用字体覆盖前，同时检查已安装模块与待生效更新。
# Inspect both installed modules and staged updates before applying overlays.
. "$MODPATH/check-conflicts.sh"
check_font_conflicts /data/adb/modules /data/adb/modules_update

# 安装包缺少必要文件时立即终止。
# Abort early if the release archive is incomplete.
[ -f "$MODPATH/system/etc/fonts.xml" ] || abort "Missing system/etc/fonts.xml"
[ -d "$MODPATH/system/fonts" ] || abort "Missing system/fonts"

# 目录需可遍历，字体与配置文件需可读取。
# Directories must be searchable and font/config files must be readable.
set_perm_recursive "$MODPATH/system" 0 0 0755 0644
