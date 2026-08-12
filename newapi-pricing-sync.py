#!/usr/bin/env python3
"""
new-api(Go 版,arcmux)批量定价脚本:以"官方价表"为锚,把模型定价为 官方价 × 系数。

适配本仓库 Go 网关的实际存储:倍率在 options 表的 ModelRatio / CompletionRatio
两个 JSON 里(不是 /api/models/ 的 model_ratio 字段)。

用法:
  python3 newapi-pricing-sync.py --base https://api.arcmux.com --token sk-xxxx --ratio 1.0 --dry-run
  python3 newapi-pricing-sync.py --base https://api.arcmux.com --token sk-xxxx --ratio 1.0

说明:
  - --ratio 0.4  = 官方价的 0.4 倍(1:1 充值下用户实付 = 官方美元价 × 0.4,单位元/百万token)
  - 分类覆盖(可选,优先级高于 --ratio):
      --override "gpt-5.6-sol:1.0,claude-opus-4.8:1.0,deepseek*:1.5"
    支持前缀匹配,deepseek* 表示所有 deepseek 开头模型。
  - 内置官方价表(2026-08-12 核对)之外的模型:不动,只警告。新增模型用
    --official-extra path.json 补官方价 {"模型名": [输入USD/百万, 输出USD/百万]}。
  - 幂等:表内模型每次从官方价重新计算,重复执行结果不变。
  - new-api 计费:输入价 = model_ratio×$2, 输出价 = model_ratio×completion_ratio×$2
    所以 ratio = 输入价/2, comp = 输出价/输入价。
"""
import argparse
import json
import sys
import urllib.request
import urllib.error

# 官方价表(USD/百万 tokens,输入/输出),2026-08-12 核对
OFFICIAL = {
    "gpt-5.6-sol":        (5.00, 30.00),
    "gpt-5.6-terra":      (2.00, 12.00),   # 7/30 官方降 20%
    "gpt-5.6-luna":       (0.20, 1.20),    # 7/30 官方降 80%
    "gpt-5.5":            (5.00, 30.00),
    "gpt-5.4":            (2.50, 15.00),
    "gpt-5.4-mini":       (0.75, 4.50),
    "gpt-5.2":            (1.75, 14.00),
    "claude-opus-4.8":    (5.00, 25.00),
    "claude-sonnet-4.6":  (3.00, 15.00),
    "claude-haiku-4.5":   (1.00, 5.00),
    "gemini-3.1-pro":     (2.00, 12.00),
    "gemini-3.6-flash":   (1.50, 7.50),
    "grok-4.5":           (2.00, 6.00),
    "deepseek-v4-pro":    (1.74, 3.48),
    "deepseek-v4-flash":  (0.14, 0.28),
}
RATIO_BASE = 2.0  # new-api: 倍率1 = $2/百万tokens
UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"


def api(base, token, method, path, body=None):
    url = base.rstrip("/") + path
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, method=method, data=data)
    req.add_header("Authorization", f"Bearer {token}")
    req.add_header("Content-Type", "application/json")
    req.add_header("User-Agent", UA)
    try:
        with urllib.request.urlopen(req, timeout=20) as r:
            return json.loads(r.read().decode())
    except urllib.error.HTTPError as e:
        print(f"HTTP {e.code} on {method} {path}: {e.read().decode()[:300]}", file=sys.stderr)
        sys.exit(1)


def parse_override(s):
    out = []
    for item in s.split(","):
        item = item.strip()
        if not item:
            continue
        name, _, ratio = item.partition(":")
        out.append((name.strip(), float(ratio)))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", required=True, help="网关地址,如 https://api.arcmux.com")
    ap.add_argument("--token", required=True, help="管理员 access token(PAT)")
    ap.add_argument("--ratio", type=float, default=1.0, help="全局系数:官方价 × N")
    ap.add_argument("--override", default="", help="分类覆盖,如 gpt-5.6-sol:1.0,deepseek*:1.5")
    ap.add_argument("--official-extra", default="", help="额外官方价表 JSON 文件,补充/覆盖内置表")
    ap.add_argument("--dry-run", action="store_true", help="只预览不提交")
    args = ap.parse_args()

    overrides = parse_override(args.override)
    official = dict(OFFICIAL)
    if args.official_extra:
        try:
            with open(args.official_extra, encoding="utf-8") as f:
                official.update(json.load(f))
        except Exception as e:
            print(f"读取 --official-extra 失败: {e}", file=sys.stderr)
            sys.exit(1)

    opts = api(args.base, args.token, "GET", "/api/option/")
    if isinstance(opts, dict) and "data" in opts:
        opts = opts["data"]
    opt_map = {it["key"]: it["value"] for it in opts}
    for key in ("ModelRatio", "CompletionRatio"):
        if key not in opt_map:
            print(f"option {key} 不存在,中止", file=sys.stderr)
            sys.exit(1)
    ratio_map = json.loads(opt_map["ModelRatio"])
    comp_map = json.loads(opt_map["CompletionRatio"])

    # 计算每个表内模型的目标值
    new_ratio, new_comp = dict(ratio_map), dict(comp_map)
    changes, fallback = [], []
    for name, (inp, out) in sorted(official.items()):
        k = args.ratio
        for pat, r in overrides:
            if name.startswith(pat.rstrip("*")) or name == pat:
                k = r
                break
        r = round(inp / RATIO_BASE * k, 6)
        c = round(out / inp, 6)
        old_r = ratio_map.get(name, "<none>")
        old_c = comp_map.get(name, "<none>")
        if old_r == r and old_c == c:
            continue
        changes.append({
            "model_name": name,
            "old": f"ratio={old_r} comp={old_c}",
            "new": f"ratio={r} comp={c}",
        })
        new_ratio[name] = r
        new_comp[name] = c

    # 表外模型提示
    for name in ratio_map:
        if name not in official and any(name.startswith(p) for p, _ in overrides):
            fallback.append(name)

    print(f"官方价表 {len(official)} 个模型,将修改 {len(changes)} 个")
    for c in changes[:40]:
        print(f"  {c['model_name']:<22} {c['old']:<28} -> {c['new']}")
    if len(changes) > 40:
        print(f"  ... 共 {len(changes)} 个")
    if fallback:
        print(f"提示: {len(fallback)} 个模型不在官方价表(未被修改): {', '.join(fallback[:20])}")

    if args.dry_run or not changes:
        print("[dry-run] 未提交" if args.dry_run else "无需修改")
        return

    api(args.base, args.token, "PUT", "/api/option/", {
        "key": "ModelRatio",
        "value": json.dumps(new_ratio, ensure_ascii=False),
    })
    api(args.base, args.token, "PUT", "/api/option/", {
        "key": "CompletionRatio",
        "value": json.dumps(new_comp, ensure_ascii=False),
    })
    print(f"已提交 {len(changes)} 个模型(约 1 分钟后 pricing 缓存刷新)")


if __name__ == "__main__":
    main()
