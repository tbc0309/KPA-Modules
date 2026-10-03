#!/system/bin/sh

MODDIR=${0%/*}

# 用原生控制进程替换 Shell，避免额外常驻 Shell。
# Replace the shell with the native controller to avoid an extra resident shell.
exec "$MODDIR/bin/kpa_rgb_daemon"
