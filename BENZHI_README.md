# BENZHI 评测说明

基于 Go 实现的民族织物纹样工艺关系复核后端服务，一款后端服务，完成织物样本与纹样单元导入、拓扑比较与工艺验证、冲突复核与反证管理、关系版本发布与冻结。

## 启动

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/task266-textilerelation --addr :8080 --db textile.db
```

## 自检（不启动长驻服务）

```bash
go run ./cmd/task266-textilerelation --smoke-test
```

`--smoke-test` 会真实创建批次与三个织物样本、解析纹样拓扑、比较 A-B 判出视觉巧合冲突、复核否决、确认 B-C 工艺传承、执行出处循环检测、发布并冻结关系版本，关闭并重开数据库验证持久化与重启恢复，最后以 0 退出码结束。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/task266-textilerelation --smoke-test
```

## HTTP API（前缀 /api）

- 批次：POST/GET /api/batches、GET /api/batches/{id}、POST /api/batches/{id}/advance、POST /api/batches/{id}/seal
- 样本：POST/GET /api/samples、GET /api/samples/{id}
- 单元：POST /api/samples/{id}/motifs、POST /api/motifs/{id}/parse、POST /api/motifs/{id}/exclude
- 工艺：PUT/GET /api/samples/{id}/technique、POST /api/samples/{id}/verify-technique
- 关系：POST /api/motifs/{from}/compare/{to}、GET /api/relations、POST /api/relations/{id}/verdict、POST /api/relations/{id}/resolve-conflict
- 反证：POST/GET /api/relations/{id}/evidence
- 版本：POST /api/versions、POST /api/versions/{id}/share、/freeze、/supersede
- 其他：GET /api/health、/api/stats、/api/selfcheck

## 持久化

SQLite（modernc.org/sqlite，CGO 无关）。七表：batches、samples、motifs、techniques、relations、counter_evidence、relation_versions。样本 sha256 幂等，关系 (from,to) 唯一，冻结版本不可变，重启同一数据库可恢复全部状态。
