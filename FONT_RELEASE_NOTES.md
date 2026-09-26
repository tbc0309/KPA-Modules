## 中文

新增安装时字体覆盖冲突检测，检查已安装模块及待生效更新；发现冲突时停止安装并提示模块 ID。允许本模块升级，跳过已停用、待移除及不自动挂载的模块。字体显示方式保持不变。

检测覆盖常见字体目录及字体配置文件，不保证识别脚本动态挂载的字体修改。

仅在 0828 固件测试，不保证通过 Play Integrity 认证。

## English

Added installation-time font overlay conflict checks for installed modules and staged updates. Conflicts stop installation and display the module ID. Self-updates, disabled modules, pending removals and modules without automatic mounts are excluded. Font rendering is unchanged.

Checks cover common font directories and configuration files, not every script-driven font modification.

Tested on firmware 0828 only. Passing Play Integrity is not guaranteed.
