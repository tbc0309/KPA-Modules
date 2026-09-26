KPA RGB Control / KPA RGB 控制
=============================

The RGB light turns off while the screen is off unless Android owns a battery indication.
屏幕熄灭时 RGB 灯关闭；Android 原厂电池提示状态除外。

电池状态每 5 秒、CPU/电池温度每 10 秒、性能模式每 2 秒检查一次。
Battery state is checked every 5 seconds, CPU/battery temperature every 10 seconds, and performance mode every 2 seconds.

性能模式 / Performance modes
----------------------------
省电 (fanMode=0)：翠绿色慢速呼吸 / Emerald slow breathing
均衡 (fanMode=1)：平滑七彩循环 / Smooth rainbow cycle
游戏 (fanMode=2)：电光蓝慢速呼吸 / Electric-blue slow breathing
火力全开 (fanMode=3)：洋红色快速脉冲 / Magenta fast pulse

优先级 / Priority
-----------------
危险温度 > 温度警告 > Android 系统电池灯 > 性能模式
Dangerous temperature > temperature warning > Android battery LED > performance mode

电量与温度 / Battery and temperature
-------------------------------------
充电、充满或 15% 以下：由 Android 控制原厂电池灯
Charging, full, or 15% and below: stock battery LED controlled by Android
CPU 85C 或电池 50C 以上：橙色快速呼吸
CPU 85C or battery 50C and above: orange fast breathing
CPU 90C 或电池 55C 以上：红色快闪
CPU 90C or battery 55C and above: red fast blink

模块只读取温度，不读取或修改风扇设置。
The module only reads temperatures and never reads or changes fan settings.

进入系统电池提示时，模块只恢复一次原厂颜色，随后停止写灯。
On battery handoff, the module restores the stock color once and then stops writing the LED.

配置文件 / Configuration
------------------------
/data/adb/kpa_rgb_control.conf

修改配置后，请停用再启用模块，或者重启设备。
After changing the configuration, disable and enable the module, or reboot.

状态与日志 / Status and log
--------------------------
/data/adb/kpa_rgb_control.state
/data/adb/kpa_rgb_control.log

仅记录状态变化与必要错误；超过 16 KB 时自动保留最近约 8 KB。
Only state changes and essential errors are logged; above 16 KB, only the newest approximately 8 KB is retained.

卸载 / Uninstall
----------------
在 Magisk 中卸载模块并重启。模块配置、状态和日志会一并清除。
Remove the module in Magisk and reboot. Its configuration, state and log are removed as well.
