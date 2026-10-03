## KPA Sitronix Fix 1.0.0

- 驱动初始化时有限重试，日志只保留本次与上次启动记录。
- Retries during driver initialization within a bounded startup window and retains only the current and previous startup logs.
- Added Magisk module update detection.
- Pauses the Sitronix health-monitor thread while preserving normal touchscreen input.
