# BW03_20260828 Sitronix 息屏耗电分析

## 结论

KPA 助手、KPA Root Helper、KPA RGB Control、KPA Touch Guard 和电脑端 ADB 均不是本次异常的来源。异常来自内核自带的 Sitronix 触摸驱动健康监控线程 `Sitronix Monitor`：息屏后线程仍在反复执行 I²C 状态检查，在本机约占用半个 CPU 核心。

驱动已经提供 `/sys/sitronix_ts_attrs/stmt` 控制节点。写入 `0` 只设置监控暂停标志，不会卸载触摸驱动或关闭触摸输入。实机写入后，65 秒息屏测试中该线程的内核 CPU 时间增量由约 `3062 tick` 降为 `0 tick`。

重新线刷至纯净 0730、保持 Bootloader 锁定且不 Root 后，连续完成两次官方 OTA 并按相同条件测试：

| 系统版本 | 活动槽 | 65 秒息屏增量 | 约占单核 |
|---|---:|---:|---:|
| 0730 | A | 2570 tick | 39.5% |
| 0813 | B | 3067 tick | 47.2% |
| 0828 | A | 3062 tick | 47.1% |

这证明异常在纯净 0730 已存在，并非 Root、KPA 助手或 Magisk 模块引入。安装修复后，连续 5 分钟息屏及三轮“锁屏 45 秒 → 唤醒解锁”测试中线程增量始终为 `0`，驱动状态始终为 `is_pause_mt = true`，触摸设备保持启用。

## 排除结果

- 卸载 KPA 助手和 KPA Root Helper、等待 Helper 自清理并重启后，异常仍存在。
- 禁用 KPA RGB Control 并重启后，异常仍存在。
- Dolby、字体模块分别禁用测试后，异常仍存在。
- KPA Touch Guard 从未安装；设备中没有其进程或残留服务。
- 电脑端停止 ADB、设备保持 `Asleep` 时，异常仍存在。
- 暂停 Sitronix 监控后，线程 CPU 增量立即归零。

## 0828 OTA 实际变更

官方更新说明包含“触摸不准确优化”，但增量包分析没有发现 Sitronix 驱动代码更新：

- 0813 与 0828 的 `sitronix_ts_monitor_thread_v3`、挂起、恢复和掉电函数机器码完全一致。
- 两版解压内核共 30,777,344 字节，仅 53 字节不同，内容为构建时间、Build ID 和 initramfs 元数据；没有触摸驱动代码或固件变化。
- DTBO 二进制哈希不同，但反编译后的设备树文本完全一致，Sitronix 节点没有参数变化。
- boot ramdisk 只变更版本属性和 SELinux 策略。
- vendor 分区只变更版本属性、NOTICE 和 SELinux 预编译策略，没有触摸固件或触摸 HAL 变化。
- system 分区涉及 `vold`、NetworkStack、Settings、services 和 SELinux；这同时对应 exFAT 修复、系统版本更新和预编译产物更新。未发现能解释内核监控线程持续耗时的驱动替换。

因此，0828 OTA 与“触摸优化”有关，但现有证据不支持“0828 修改了 Sitronix 监控线程并引入耗电”的说法。更准确的判断是：原有驱动监控逻辑在当前触控状态下持续执行耗时的控制器检查；OTA 可能改变了控制器运行状态或让原有问题更容易出现，但不是 KPA 模块残留造成的。

## 临时修复

Root shell 可执行：

```sh
echo 0 > /sys/sitronix_ts_attrs/stmt
```

验证：

```sh
cat /sys/sitronix_ts_attrs/stmt
# is_pause_mt = true
```

恢复监控：

```sh
echo 1 > /sys/sitronix_ts_attrs/stmt
```

重启也会恢复驱动默认值。

## 模块修复

`KPA Sitronix Fix` 在开机后等待控制节点出现，写入 `0` 并校验状态，然后退出，不驻留、不轮询、不持有唤醒锁。模块卸载脚本会写入 `1` 恢复监控；设备不存在该节点时不会修改任何内容。

暂停的是驱动的健康检查与自动复位机制。正常触摸输入路径保持启用，但如果触控控制器以后真实失去响应，驱动不会再由该监控线程自动复位；此时重启设备或卸载模块即可恢复默认行为。根本修复仍应由厂商修改驱动，使监控在面板休眠时停止，并避免 I²C 异常路径长时间占用 CPU。
