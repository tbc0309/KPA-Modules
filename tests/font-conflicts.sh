#!/system/bin/sh
set -eu
. "${1:?Path to check-conflicts.sh required}"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
ui_print() { echo "$*"; }
abort() { echo "$*"; exit 1; }
expect() {
  expected=$1
  label=$2
  if (check_font_conflicts "$work/modules" "$work/updates"); then result=0; else result=1; fi
  [ "$result" = "$expected" ] || { echo "FAIL: $label"; exit 1; }
  echo "PASS: $label"
}
mkdir -p "$work/modules/kpa_myuppy_font/system/fonts" "$work/updates"
expect 0 self-update
mkdir -p "$work/modules/other/system/fonts"
expect 1 active-font
touch "$work/modules/other/disable"
expect 0 disabled-font
mv "$work/modules/other/disable" "$work/modules/other/remove"
expect 0 pending-removal
mv "$work/modules/other/remove" "$work/modules/other/skip_mount"
expect 0 no-automatic-mount
mkdir -p "$work/updates/staged/system/product/etc"
touch "$work/updates/staged/system/product/etc/fonts_customization.xml"
expect 1 staged-configuration
touch "$work/updates/staged/disable"
mkdir -p "$work/modules/audio/system/lib64"
expect 0 unrelated-module
