#!/system/bin/sh

MODDIR=${0%/*}

# Replace the shell with the low-power native controller.
exec "$MODDIR/bin/kpa_rgb_daemon"
