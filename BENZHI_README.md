基于 Go 实现的民族织物纹样工艺关系复核 Web 项目，一款后端服务，导入织物样本与纹样单元、比较单元拓扑相似度、验证经纬交错规则与染料证据并发布不可变工艺传承关系版本。

# BENZHI 评测说明

民族织物纹样工艺关系复核台（纯后端）。

## 运行契约

- **启动服务**：`/app/textilerelation --addr :8080 --db textile.db`
- **端到端自检**：`/app/textilerelation --smoke-test --db smoke.db`
  - 真实创建批次与三个织物样本（平纹 1/1 / 斜纹 3/1 / 斜纹 3/1）及纹样单元，
    解析单元拓扑（周期+对称）、写入工艺特征、比较 A-B 判出"外观相似但交错规则相反 → 视觉巧合"、
    复核否决、确认 B-C 工艺传承、执行出处循环检测、发布并冻结关系版本，
    关闭并重开同一数据库验证持久化与重启恢复，最终以退出码 0 结束。
  - 这是 Docker `CMD` 与双架构验证的唯一判据，**只传 flag，不传路径位置参数**。

## Docker 双架构验证

```bash
# amd64
docker buildx build --platform linux/amd64 --load -t textilerelation:amd64 .
docker run --rm textilerelation:amd64 --smoke-test

# arm64
docker buildx build --platform linux/arm64 --load -t textilerelation:arm64 .
docker run --rm textilerelation:arm64 --smoke-test
```

## API 一览（前缀 /api）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/batches | 新建样本批次 |
| GET | /api/batches | 列出批次 |
| GET | /api/batches/{id} | 批次详情 |
| POST | /api/batches/{id}/advance | 推进批次状态机 |
| POST | /api/batches/{id}/seal | 封存批次 |
| POST | /api/samples | 导入织物样本（哈希幂等） |
| GET | /api/samples | 列出样本（?batch_id= 过滤） |
| GET | /api/samples/{id} | 样本详情 |
| POST | /api/samples/{id}/motifs | 添加纹样单元 |
| GET | /api/samples/{id}/motifs | 列出样本单元 |
| GET | /api/motifs | 列出全部单元 |
| GET | /api/motifs/{id} | 单元详情 |
| POST | /api/motifs/{id}/parse | 解析单元拓扑（周期+对称） |
| POST | /api/motifs/{id}/exclude | 排除单元（?patch=true 标记补片） |
| PUT | /api/samples/{id}/technique | 写入工艺特征 |
| GET | /api/samples/{id}/technique | 查询工艺特征 |
| GET | /api/techniques | 列出工艺特征 |
| POST | /api/samples/{id}/verify-technique | 织造核验 |
| POST | /api/motifs/{from}/compare/{to} | 比较两单元生成候选 |
| GET | /api/relations | 列出候选（?verdict= 过滤） |
| GET | /api/relations/{id} | 候选详情（含反证） |
| POST | /api/relations/{id}/verdict | 人工裁决 |
| POST | /api/relations/{id}/resolve-conflict | 复核冲突候选 |
| POST | /api/relations/{id}/evidence | 追加反证 |
| GET | /api/relations/{id}/evidence | 列出反证 |
| POST | /api/versions | 创建关系版本草稿 |
| GET | /api/versions | 列出版本 |
| GET | /api/versions/{id} | 版本详情（含关系） |
| POST | /api/versions/{id}/share | 共享版本 |
| POST | /api/versions/{id}/freeze | 冻结版本 |
| POST | /api/versions/{id}/supersede | 用新版本替代 |
| GET | /api/health | 健康检查 |
| GET | /api/stats | 汇总统计 |
| GET | /api/selfcheck | 自检 |

## 示例

```bash
# 建批次
curl -X POST localhost:8080/api/batches -d '{"name":"b1","note":"样例批次"}'

# 导入样本（幂等：重复导入返回既有样本）
curl -X POST localhost:8080/api/samples -d '{
  "batch_id":1,"name":"fragment-a","provenance":"圣但尼修道院",
  "warp_count":12,"weft_count":12,"warp_density":18,"weft_density":16,
  "motif_grids":{"diamond-a":"##.#,.##.,##.#,.##."}
}'

# 写入工艺特征
curl -X PUT localhost:8080/api/samples/1/technique -d '{
  "weave_class":"plain","interlacing":"1/1","dye_class":"natural",
  "dye_pigment":"茜草","colorfastness":4,"carbon_ratio":-18.5
}'

# 解析单元并比较
curl -X POST localhost:8080/api/motifs/1/parse
curl -X POST localhost:8080/api/motifs/1/compare/2

# 追加出处反证（自动否决候选）
curl -X POST localhost:8080/api/relations/1/evidence -d '{
  "kind":"opposing_source","description":"两样本年代差 300 年","ref":"档案-12"
}'

# 发布版本
curl -X POST localhost:8080/api/versions -d '{"name":"v1","summary":"传承关系确认"}'
curl -X POST localhost:8080/api/versions/1/share
curl -X POST localhost:8080/api/versions/1/freeze
```
