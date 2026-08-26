# task266-textilerelation 民族织物纹样工艺关系复核台

面向纺织史研究者的织物纹样工艺传承证据复核服务。研究者导入织物样本、纹样单元、经纬结构与染料证据，
系统比较纹样单元拓扑、验证经纬交错规则与染料兼容性，生成工艺传承关系候选；
研究者可排除后期补片、添加出处反证并发布冻结的关系版本。

## 业务闭环

1. **导入**：创建批次 → 导入织物样本（内容哈希幂等）→ 录入纹样单元；
2. **解析**：解析单元拓扑（平铺周期、对称类型）并写入工艺特征（织法/浮长/染料）；
3. **比较**：比较单元拓扑相似度，结合经纬交错规则与染料证据生成传承候选；
4. **复核**：裁决候选、处理"外观相似但交错规则相反"的视觉巧合、追加反证；
5. **发布**：创建关系版本（出处循环检测）→ 共享 → 冻结不可变快照 → 可被新版本替代。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...

# 端到端自检（创建真实数据 → 关闭重开数据库验证持久化 → 退出码 0）
go run ./cmd/task266-textilerelation --smoke-test --db smoke.db

# 启动服务
go run ./cmd/task266-textilerelation --addr :8080 --db textile.db
```

## API 入口

全部路由以 `/api` 为前缀，详见 BENZHI_README.md 的 API 一览表。
核心入口示例：`POST /api/samples`（幂等导入）、`POST /api/motifs/{from}/compare/{to}`（候选生成）、
`POST /api/relations/{id}/evidence`（反证）、`POST /api/versions/{id}/freeze`（冻结发布）。

## 持久化

SQLite（modernc.org/sqlite 纯 Go 驱动，CGO 无关）：batches / samples / motifs / techniques /
relations / counter_evidence / relation_versions 七张表，关闭后重开同一数据库即可完整恢复，
`--smoke-test` 全程验证该重启恢复路径。
