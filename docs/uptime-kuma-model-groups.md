# 按模型分组接入 Uptime Kuma

本配置为 `chatgpt`、`claude`、`grok` 各探测一个代表模型。模型和分组来自 2026-09-13 的线上 `/api/pricing`；尚未使用生产密钥验证下面的推理请求。监控成功表示所选模型通过指定分组的调用成功，不代表整组所有模型、所有渠道均正常。

## 1. 在 ArcMux 创建监控专用 API Key

进入「API 密钥」，分别创建：

| 密钥名称 | 固定分组 | 如启用模型限制，只允许 |
| --- | --- | --- |
| uptime-chatgpt | chatgpt | gpt-5.5 |
| uptime-claude | claude | claude-opus-4-6 |
| uptime-grok | grok | grok-4.5 |

不要选择自动分组，关闭跨分组重试。确保创建密钥的账户有对应分组权限、可用余额，密钥未过期且有可用额度。密钥仅填入 Kuma 的私有监控配置，不要放进公开状态页描述。

## 2. 在 Kuma 添加三个监控

点击「Add New Monitor」，三个监控使用以下共同配置：

| 字段 | 值 |
| --- | --- |
| Monitor Type | HTTP(s) - JSON Query |
| URL | https://api.arcmux.com/v1/chat/completions |
| Heartbeat Interval | 300 秒 |
| Retries | 2 |
| Retry Interval | 60 秒 |
| Request Timeout | 60 秒 |
| Accepted Status Codes | 200 |
| Method（HTTP Options 中） | POST |
| Body Encoding | JSON |
| Authentication | None（认证放在 Headers） |
| JSON Query | `$exists(choices[0].message) and $not($exists(error))` |
| 比较条件 | `==` |
| Expected Value | `true`（不加引号） |

每个监控的 Headers 填以下 JSON，并将占位符替换为该分组的完整密钥：

```json
{
  "Authorization": "Bearer <该分组的完整 API Key>",
  "Content-Type": "application/json"
}
```

### OpenAI 代表模型

名称：`OpenAI · gpt-5.5`。Headers 使用 `uptime-chatgpt` 密钥。

Body：

```json
{
  "model": "gpt-5.5",
  "messages": [{ "role": "user", "content": "Reply only OK." }],
  "max_completion_tokens": 128,
  "stream": false
}
```

### Claude 代表模型

名称：`Claude · claude-opus-4-6`。Headers 使用 `uptime-claude` 密钥。

Body：

```json
{
  "model": "claude-opus-4-6",
  "messages": [{ "role": "user", "content": "Reply only OK." }],
  "max_tokens": 32,
  "stream": false
}
```

### Grok 代表模型

名称：`Grok · grok-4.5`。Headers 使用 `uptime-grok` 密钥。

Body：

```json
{
  "model": "grok-4.5",
  "messages": [{ "role": "user", "content": "Reply only OK." }],
  "max_tokens": 32,
  "stream": false
}
```

先确认每个监控至少完成一次成功请求。这里检查返回结构，不要求回复必须精确为 OK；推理模型可能将有限输出预算用于推理。HTTP 200 且存在 `choices[0].message`、没有顶层 `error` 才算通过。

每 5 分钟检查一次，每个监控每天约 288 次，三个合计约 864 次，失败重试另计。请求正常计费。若返回 400，先核对上游参数支持；401/403 检查密钥与分组权限；429 检查额度、限流和上游限制，不能直接认定整组模型宕机。

## 3. 加入现有公开状态页

在 Kuma 的「Status Pages」打开 `arcmux` 页面，点击「Edit Status Page」：

1. 建立 OpenAI、Claude、Grok 三个展示分组。
2. 各组加入对应的代表模型监控。
3. 原有的网关监控可以单独放在「平台服务」组；核对其 URL，监测 Kuma 自身与监测 ArcMux 网关含义不同。
4. 保存，用无痕窗口确认三个模型监控都可见。

## 4. ArcMux 仍然只接入一个状态页

在「系统设置 → 内容 → Uptime Kuma」编辑现有条目：

| 字段 | 值 |
| --- | --- |
| 分类名称 | 模型服务 |
| Uptime Kuma URL | https://status.arcmux.com |
| 状态页面 Slug | arcmux |

点击「更新」，再点击外层「保存设置」。不要为同一 URL 和 Slug 建三个 ArcMux 条目，否则会重复显示同一状态页的全部监控。

ArcMux 将展示三个独立监控及各自的近 24 小时可用率。历史柱条需要同时部署支持 `heartbeats` 的后端和前端；旧版本只有状态点与百分比。

## 核对依据

- ArcMux：`middleware/auth.go`（密钥分组）、`controller/uptime_kuma.go`（状态页聚合）、`relay/channel/openai/adaptor.go`（GPT-5 参数转换）。
- Kuma 2.0.0：[监控表单](https://github.com/louislam/uptime-kuma/blob/2.0.0/src/pages/EditMonitor.vue)、[JSON Query 比较逻辑](https://github.com/louislam/uptime-kuma/blob/2.0.0/src/util.ts)。
