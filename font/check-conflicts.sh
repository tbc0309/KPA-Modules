#!/system/bin/sh

check_font_conflicts() {
  for root in "$@"; do
    for module in "$root"/*; do
      [ -d "$module" ] || continue
      # Updating this module is safe; disabled/removed modules do not mount next boot.
      [ "${module##*/}" = kpa_myuppy_font ] && continue
      [ -f "$module/disable" ] && continue
      [ -f "$module/remove" ] && continue
      [ -f "$module/skip_mount" ] && continue
      for partition in system system/product system/vendor system/system_ext product vendor system_ext; do
        base="$module/$partition"
        if [ -d "$base/fonts" ] || [ -f "$base/etc/fonts.xml" ] ||
           [ -f "$base/etc/font_fallback.xml" ] || [ -f "$base/etc/fonts_customization.xml" ]; then
          ui_print "! Conflicting font overlay: ${module##*/}"
          abort "Disable or remove this module, then retry. Reboot after installation."
        fi
      done
    done
  done
}
