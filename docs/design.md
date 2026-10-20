# lecoo-control-gui 设计方案

## 1. 产品边界

给 N155A 这台机器补一个图形界面。所有能力来自 `C:\Program Files\LecooControlCenter\lecoo-ctrl.exe` 的 12 个子命令，本项目只用到其中与温度、风扇、电源、充电相关的部分。界面不直接访问嵌入式控制器，不加载驱动，不替换守护进程。

这台机器的现状是只有底座没有门面。`LecooControlDaemon` 服务在跑（`lecoo-ec-daemon.exe --service`，自动启动），驱动 `inpoutx64` 已加载，命令行可用，但没有图形界面。用户要看温度和转速，并且点一下切换风扇、电源、充电。

范围之外:

- EC 寄存器直读，PWM 波形，传感器解码
- 键盘背光与后置灯环，即 `kbd` 与 `led` 两个命令
- CLI 没有的功能
- 非 N155A 机型适配
- 守护进程本身的修改或替换

### 底座进程

`lecoo-ec-daemon.exe --service` 是以 LocalSystem 身份自动启动的 Windows 服务，服务名 `LecooControlDaemon`，显示名 `Lecoo EC Daemon`，描述为 `Lecoo laptop EC hardware control daemon`。它是这台机器上唯一接触嵌入式控制器的进程，通过 `inpoutx64` 内核驱动读写 IT5571 的 EC 寄存器，日志中的 `Board N155A, chip IT5571-07, HRAM window 0x0400` 是它启动时的探测结果。常驻开销很小，实测内存 9.93 MB、7 个线程、累计 CPU 时间 0.08 秒。

CLI 不碰硬件。`lecoo-ctrl.exe` 的字符串里没有 `inpout`、`DeviceIoControl`、`IT5571`，只有命名管道和一句 `Failed to connect to daemon! Is it running?`。整条链路是客户端打开 `\\.\pipe\lecoo_ctl_daemon` 发送 JSON，守护进程负责翻译成寄存器操作。

守护进程的接口是一组命令枚举，包括 `GetCapabilities`、`GetSettings`、`ApplySettings`、`RestoreDefaults`、`GetTelemetryId`、`ActivateTelemetry`，以及关机、休眠、唤醒三个准备钩子。读数与设置是另一层请求，涵盖温度、转速、充电状态、充电限制、电源配置、键盘背光与 LED。`--help` 末尾那段 `Hardware Capabilities (Board: N155A)` 由它提供，守护进程不在时该段消失。

两件界面必须依赖但看不见的行为。第一，它持久化状态，格式为 JSON，路径 `C:\ProgramData\LecooControl\daemon_state.bin`，客户端断开时保存。实测把电源改成 `Silent`、风扇改成 `Custom(140)` 后重启 `LecooControlDaemon` 服务，两者自行恢复，因此界面不需要自己做持久化，也不需要启动时重发。第二，充电限制是状态机而非单次寄存器写入，二进制中含 `Charge policy pending:`、`charge reconcile on AC connect:`、`Will arm once the battery drops to`、`Requires active charging: connect the AC adapter`。这正是 `charge` 同时打印目标与生效的原因。

遥测默认开启，由守护进程而非 CLI 发出。目标 `https://lab.lavashik.dev/telemetry/v2`，请求头带 `X-Daemon-Version`，JSON 体包含从注册表读出的 `vendor`、`product`、`motherboard`、`bios`、`cpu`、`os`、`arch`，以及实时温度、转速、风扇模式、背光、LED、充电、`soc`、`uptime_s`、`Startup` 事件和转发的错误日志。`daemon telemetry disable` 可关闭，关闭不影响其他功能。

## 2. 一个约束决定了整体架构

写入命令只返回一行 `完成。` 和退出码 0。这一行只证明命令被受理，不证明硬件接受了设置。读回同样状态时 CLI 给出三种互不相同的格式:

```
$ lecoo-ctrl.exe power silent        # 写
完成。

$ lecoo-ctrl.exe power               # 读方式一
⚡ 电源配置:
   当前: Silent

$ lecoo-ctrl.exe daemon settings read  # 读方式二
CurrentSettings {
    power_profile: Silent,
    charge: Full,
    fan_mode_cpu: Auto,
    ... }
```

充电更能说明问题。同一个状态，`charge` 给的是中文人话 `保护模式 (40% .. 80%)`，`daemon settings read` 给的是嵌套结构 `Preserve(Some(ChargeRange{min,max}))`。前者只能靠正则猜，后者可以直接解析。

结论有两条。第一，界面渲染的状态必须来自写完之后的一次真实读回，不能乐观改成新值。第二，离散设置只认 `daemon settings read` 一个来源，其余读回格式不作为状态来源。

`daemon settings read` 是唯一一处英文、无 emoji、结构稳定、一次覆盖全部离散项的输出。它采用 Rust `Debug` 美化格式，不是 JSON，但有确定的嵌套形状。下面省略了本项目不使用的 `keyboard_backlight` 与 `led_mode` 两行:

```
CurrentSettings {
    telemetry_enabled: true,
    telemetry_client_id: 1003880566478677932,
    power_profile: Default,
    charge: Preserve(
        Some(
            ChargeRange {
                min: 40,
                max: 80,
            },
        ),
    ),
    fan_mode_cpu: Custom(
        120,
    ),
    fan_mode_gpu: Auto,
}
```

三种形态要一起处理。裸枚举 `Auto`，跨行元组枚举 `Custom(120,)`，跨行嵌套可选结构 `Preserve(Some(ChargeRange{min,max}))`。

## 3. CLI 面参考

实测于 v0.5.2-beta，主板 N155A。所有命令输出为无 BOM 的 UTF-8，成功时退出码 0，失败时退出码 1 且错误写 stderr。键盘背光 `kbd` 与后置灯环 `led` 不在本项目范围内，不列入。

| 命令 | 读回 | 写入 | 说明 |
| --- | --- | --- | --- |
| `info` | 控制器型号、控制器版本、HRAM 偏移、守护进程版本 | 否 | 冷启动时取一次 |
| `temps` | CPU 温度、系统温度 | 否 | 文本含 emoji 与中文 |
| `fans` | CPU 转速、GPU 转速 | 否 | 单位 RPM |
| `monitoring --rate N` | 持续输出同一行数据 | 否 | 实时数据唯一来源，样本用 `\r` 分隔 |
| `fan <cpu\|gpu\|both> <auto\|full\|turbo\|custom> [pwm]` | 无 | 是 | `custom` 必须带 0 到 255 |
| `charge [preset\|--min N --max N]` | 目标、生效、电池百分比 | 是 | 预设 full、desk、lifespan、balanced、high；自定义范围 20 到 100 |
| `power [silent\|default\|perf]` | 当前配置 | 是 | 读回显示 Silent、Default、Performance |
| `daemon telemetry enable\|disable\|id` | 遥测 ID | enable、disable | |
| `daemon settings read\|apply\|reset` | 完整设置结构 | apply 重新下发，reset 恢复默认 | |
| `daemon version` | 同 `info` | 否 | |
| `hwtest` | 交互式向导 | 是 | 需要真实终端，无法通过管道驱动 |
| `--version` | `lecoo-ctrl.exe v0.5.2-beta` | 否 | 用于版本闸门 |

`temps` 单条约 8 毫秒 CPU、16 毫秒墙钟，`fans` 约 7 毫秒 CPU、14 毫秒墙钟，`info` 18 毫秒。并发 6 个 `fans` 进程同时读取全部成功，没有观察到互斥或锁冲突。单次调用的成本几乎全在创建进程，而不是读 EC。

## 4. 分层

```
UI (页面与控件)
  │  只读类型化状态, 只发类型化意图
  ▼
Store (状态聚合, 采样调度)
  │
  ▼
Gateway (一个命令一个函数, 参数用数组传)
  │
  ▼
Process (spawn, 隐藏窗口, 按 UTF-8 解码, 超时与退出码)
  │
  ▼
Parser (唯一的文本边界)
  │
  ▼
lecoo-ctrl.exe
```

解析全部集中在 Parser 一层。Gateway 之上没有任何字符串处理，UI 拿到的永远是类型。这是边界纪律的具体落地，也让解析规则可以单独测试，不需要启动界面。

Gateway 每个函数返回一个结果类型，不抛异常给 UI:

```ts
type CliResult<T> = { ok: true; value: T } | { ok: false; error: CliError };

type CliError =
  | { kind: 'missing-cli'; path: string }
  | { kind: 'daemon-down' }
  | { kind: 'rejected'; message: string }   // exit 1, error: 行
  | { kind: 'timeout' }
  | { kind: 'unparsed'; raw: string };
```

## 5. 状态模型

离散状态用一个联合类型表达，和 `daemon settings read` 一一对应，不做多余的中间表示:

```ts
type Pwm = number;                     // 0..255

type FanMode =
  | { kind: 'auto' }
  | { kind: 'full' }
  | { kind: 'custom'; pwm: Pwm }
  | { kind: 'turbo' };

type PowerProfile = 'silent' | 'default' | 'perf';

type ChargeTarget =
  | { kind: 'preset'; preset: 'full' | 'desk' | 'lifespan' | 'balanced' | 'high' }
  | { kind: 'preserve'; min: number | null; max: number | null }
  | { kind: 'unknown'; raw: string };

type DaemonSettings = {
  telemetryEnabled: boolean;
  telemetryClientId: string;           // u64 超过 JS 安全整数, 按字符串存
  power: PowerProfile | { kind: 'unknown'; raw: string };
  charge: ChargeTarget;
  fanCpu: FanMode;
  fanGpu: FanMode;
  readAt: number;
};
```

遥测和电池是另一种形状，带时间戳，只用于展示和曲线:

```ts
type Sample = {
  at: number;
  cpuTempC: number;
  sysTempC: number;
  cpuRpm: number;
  gpuRpm: number;
};

type BatteryState = {
  target: string;      // 目标, 原样保留可读文本
  effective: string;   // 生效
  percent: number;
  at: number;
};
```

`DaemonSettings` 是唯一权威状态。界面上的每个开关都从这个对象渲染。写操作完成后立即重新读一次，用新对象替换旧对象，界面被动更新。这样就不存在"界面显示已开启但硬件没接受"的状态。

`ChargeTarget` 保留 `unknown` 分支，因为 CLI 版本变化时可能出现没见过的枚举。未知值原样显示，界面不崩。

电池的 `target` 与 `effective` 分开显示。CLI 明确给出这两个字段，说明它们可能不一致，界面必须在两者不同时提示"已下发，等待生效"。

## 6. 实时数据与刷新节奏

实时数据只走一条路，`monitoring --rate 1` 常驻一个子进程，不分别轮询 `temps` 和 `fans`。这是实测决定的，不是偏好。

| 方式 | 每样本 CPU | 组成 |
| --- | --- | --- |
| `temps` 加 `fans` 轮询 | 15.1 毫秒 | `temps` 8.33 毫秒加 `fans` 6.77 毫秒，几乎全花在 CreateProcess 与镜像加载 |
| `monitoring --rate 1` | 测不出，低于计时精度 | 只创建一次进程，之后复用 |

同一套测量方法（`CreateNoWindow`、重定向 stdout、读 `TotalProcessorTime`）。按 1 Hz 持续采样，轮询方式每小时烧掉 54.4 秒 CPU，占单核 1.5%。`monitoring` 在 10 Hz 下跑 23 秒，CPU 读数仍为 0。守护进程在两种方式下都读到 0，成本全在客户端，而贵的是每次都创建进程，不是读 EC。

因此 `monitoring` 独占实时通道，提供 CPU 温度、系统温度、CPU 转速、GPU 转速。它只输出文本和 `\r`，没有 ANSI 控制码，样本之间用回车分隔，适合按行解析。其余数据各有节奏。

| 数据 | 来源 | 窗口前台 | 隐藏到托盘 |
| --- | --- | --- | --- |
| 温度与转速 | `monitoring` | `--rate 1` | `--rate 5` |
| 电池百分比、目标、生效 | `charge` | 启动，写完，每 30 秒 | 每 60 秒 |
| 离散设置 | `daemon settings read` | 启动，写完，获得焦点，每 30 秒 | 每 120 秒 |
| 控制器信息 | `info` | 启动一次 | 不轮询 |

速率只能在启动时指定，所以窗口显隐切换时重启子进程。子进程必须挂到 Job 对象上并设置 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`，否则界面崩溃会留下孤儿进程一直读 EC。杀掉子进程会让守护进程记一条 `Could not acquire lock to save state on connection reset`，不影响功能。

子进程意外退出时按 1 秒、2 秒、5 秒、10 秒退避重启。连续失败五次就降为轮询 `temps` 和 `fans`，并在总览页标注“实时数据已降级”。这是保留轮询解析器的唯一理由。

任何一次采样失败不清空旧值，只把面板标记为“数据陈旧”并显示最后更新时间。

## 7. 界面结构

窗口无边框，顶部 40 像素自定义标题栏承担拖动、主题切换与最小化、最大化、关闭。初始尺寸 1320x880，最小 1060x700，可缩放。左侧 188 像素导航，五项。界面支持明暗两套主题，选择存在 localStorage，首次启动默认夜间。

**总览**。四张实时卡片，CPU 温度、系统温度、CPU 风扇、GPU 风扇，每张带 120 秒迷你曲线。下面一行是电池环形百分比加目标与生效两个字段。再下面一行用 chip 显示当前电源模式与风扇模式，点击 chip 跳到对应页面。

**风扇**。顶部是目标选择，同步（both）、CPU、GPU。中间是模式分段控件，自动、全速、自定义。自定义时出现 0 到 255 的滑块加数字输入，松手才下发，拖动过程只改数字不下发。旁边显示当前实际转速，下发后 3 秒刷新一次以反映变化。全速下方有一条常驻说明，讲清噪音与功耗代价。

`turbo` 不进分段控件。它出现在页面底部的折叠区域，标注为不安全模式，点击后弹出二次确认，确认文案写明风扇将以最大速度运行。默认永不开启，不记忆。

**电源**。三张大卡片，静音、默认、性能，点击即生效，当前项高亮。卡片上写入各自取向。性能模式旁边提示它与风扇联动，自动风扇会更早升速。

**充电**。五张预设卡片，全容量 100%，桌面 40%，最大寿命 60%，平衡 80%，高容量 95%，当前项高亮。下面是自定义区间，两个滑块控制 min 与 max，范围 20 到 100，提交前校验 min 小于 max。再下面是目标、生效、电池三行实时值。桌面模式旁注明适合长期插电使用。

**关于与高级**。控制器型号、控制器版本、HRAM 偏移、CLI 版本、守护进程版本。遥测开关与遥测 ID（带复制按钮）。四个操作按钮，重新下发设置、恢复守护进程默认、打开硬件自检、设置 CLI 路径。最下面是一个只读文本框，显示最近一次 `daemon settings read` 的原文，方便排查。

**精简模式**。标题栏右侧的按钮把窗口从 1320x880 缩到 380x232，整页换成一个小部件。内容是 2x2 网格（CPU 温度、系统温度、CPU 转速、GPU 转速）加一行底栏（电池百分比、进度条、当前电源模式），再配一条 32 像素的拖动条和图钉、展开、关闭三个按钮。默认置顶，图钉可取消。展开时恢复原尺寸并居中，同时取消置顶。精简状态记在 localStorage，下次启动直接以小组件形态打开。

## 8. 危险操作闸门

| 操作 | 处理 |
| --- | --- |
| `fan ... turbo` | 折叠区加二次确认，不记忆选择 |
| `fan ... custom` | 滑块松手下发，禁止拖动中连续写 |
| `daemon settings reset` | 输入框要求键入 `RESET` 才启用按钮 |
| `daemon telemetry disable` | 直接可点，旁边解释关闭后不再上报 |
| `hwtest` | 不内嵌，用新控制台窗口启动，因为它是交互式向导 |
| `daemon settings reset` 之外的写入 | 写完自动读回，失败不保留界面上的新值 |

滑块下发统一加 150 毫秒防抖，并且只在下发成功且读回匹配后才把滑块位置锁定为状态值。

## 9. 托盘与启动

托盘图标颜色反映当前电源模式。悬停提示显示 CPU 温度与 CPU 转速。右键菜单包含电源模式子菜单、风扇模式子菜单、显示主窗口、开机自启开关、退出。左键单击切换窗口显隐。

开机自启用当前用户注册表 `Run` 项或计划任务，不需要管理员权限。单实例用互斥锁，第二次启动只唤醒已有窗口。

全局快捷键默认关闭，可在设置里开启电源模式循环切换。

## 10. 技术选型

已定并已实现：Go 加 Wails v2，前端 Svelte 5 加 shadcn-svelte、Tailwind v4、LayerChart 2 与 Motion。

| 层 | 选择 | 理由 |
| --- | --- | --- |
| 外壳 | Wails v2 | Go 侧直接 `exec` 调 CLI，前台没有 Node 运行时，产物 12 MB |
| 界面 | Svelte 5 runes | 状态集中在 `.svelte.ts`，不需要额外状态库 |
| 组件 | shadcn-svelte 加 Tailwind v4 | 组件落进仓库，颜色走 CSS 变量，明暗主题只切 `.dark` |
| 图表 | LayerChart 2 | `LineChart` 自带时间轴、刻度与高亮，省掉手写 SVG 轴 |
| 动效 | Motion | 只做入场与提示，不参与数据路径 |

放弃的备选是 Tauri 2 与 .NET 8 加 WPF。前者要在 Go 侧另写一层 Rust FFI，换不来收益；后者用 XAML 写曲线和主题切换更费力。

三条实现约束，与选型无关。

- 传参用 argv 数组，不拼接命令行字符串。
- 子进程必须隐藏控制台窗口，否则每次创建子进程都会闪黑框。
- 读原始字节并按 UTF-8 解码，不要走系统代码页转换。

## 11. 错误与降级

启动时先读 `--version`。版本与开发时记录的 `v0.5.2-beta` 不一致时，顶部显示一条黄色横幅，提示解析规则可能失效，功能照常可用。

CLI 路径不存在时，总览页替换为一段说明和"设置 CLI 路径"按钮，其余页面隐藏，避免整屏报错。

守护进程不可用时，表现为读命令失败。此时显示"守护进程未运行"并提供启动按钮，启动通过 `sc start LecooControlDaemon` 触发 UAC。这是唯一需要提权的路径。

单条解析失败时显示 `—` 而不是 0，并在关于页保留原始输出。0 和"读不到"在用户眼里必须不同。

`error:` 开头的 stderr 行做映射，覆盖取值无效、缺少必需参数、无法识别的命令三类，其余原样显示。CLI 输出为中文，解析规则优先用数字位置而不是标签文字，这样即使将来输出去掉中文标签也不至于全盘失效。

## 12. 里程碑

**M1 只读总览**。进程层、解析层、`monitoring`、`charge`、`daemon settings read`、总览页。此时能验证解析是否可靠，不涉及任何写入风险。

**M2 离散写入**。电源与充电预设，走写完读回的闭环。

**M3 风扇与自定义值**。风扇三目标与三模式，PWM 滑块，充电自定义区间，全部危险闸门。

**M4 托盘与启动**。托盘菜单、单实例、开机自启、窗口隐藏时降低采样率。

**M5 高级页与分发**。遥测、原始转储、错误降级、安装包与图标。

## 13. 未验证项

第一条已结案，其余都在实现 M1 之前应确认。

**权限需求已实测，界面无需提权。** 管道 `\\.\pipe\lecoo_ctl_daemon` 的 DACL 为 `(A;;FA;;SY)(A;;FA;;BA)(A;;0x12019f;;BU)`，其中 BU 是 Builtin Users，掩码含读写数据权限。普通用户进程可以连接守护进程，因此安装与运行都不需要 UAC。本机 UAC 已关闭（`EnableLUA` 为 0），这一点无法在本机用非提权进程直接验证，结论来自 ACL 判读而非运行测试。

**以下值从未观察到。** `fan ... turbo` 的实际读回表示，`ChargeTarget` 的 `Preserve(None)` 形态，以及设置为全速后 `fan_mode_cpu` 的确切文本。解析器必须容忍未知枚举，把原文交给界面。

**目标与生效是否真会分叉未观察到。** 界面按会分叉来设计，因此两者都显示。

**`monitoring` 的异常行为未观察到。** 守护进程重启后子进程是否自恢复、daemon 不可用时它是否报错退出，都未测。设计上靠监控子进程退出码加退避重启处置，降级到轮询。

**风扇模式会在热负载下自行变成 `Custom(n)`。** 实测 CPU 冲到 95°C 时，`fan_mode_cpu` 从 `Auto` 变为 `Custom(181)`，几秒后又变成 `Custom(170)`，期间没有任何客户端写入。执行 `fan both auto` 还原后连续 84 秒保持 `Auto`。推断是 EC 或守护进程的风扇爬坡把当前占空比写回了设置，尚未证实。界面目前按真实读回展示，但风扇页不应把读到的 `Custom` 当成用户选的自定义值，否则滑块会被硬件跳变带走。这一条需要先定下行为再改风扇页。

**CLI 没有 JSON 输出，也没有稳定契约。** 整个设计用版本闸门加解析失败降级来吸收变化,而不是依赖 CLI 承诺。
