# 无障碍文档整改验收台

本项目面向公共服务内容编辑与无障碍质量审核员，用于把一份待发布文档从建档、结构化内容录入、规则审查、逐项整改和人工复核推进到合格声明发布。系统保留只追加事件时间线，并使用聚合 `revision` 实施乐观并发控制、使用 `request_id` 保证写操作幂等。

服务为单进程 Go 应用，不依赖外部数据库或 Node 构建链。浏览器工作台、CSS 和 JavaScript 均由 Go 服务直接托管。

## 主要能力

- 创建与维护验收案元数据，保存标题、机构、负责人和目标发布日期。
- 按标题或机构检索，并按负责人、状态和目标发布日期组合筛选；支持稳定游标分页、到期排序和逾期提示。
- 录入标题、段落、链接、图片和表格等结构化内容，保留不可变内容修订历史，支持版本差异和安全回退。
- 确定性检查标题跳级、空链接文本、缺失图片替代文本和表格表头缺失，并在内容更新后进行增量复审和问题差异追踪。
- 逐项或批量提交绑定最新内容修订的整改证据，批量操作按单一聚合修订原子生效。
- 由审核员逐项或批量接受、退回证据，仅在全部最新证据通过后允许预览和批准声明。
- 批准时锁定规范化声明及稳定摘要，发布前校验完整性；打印页面和 JSON 导出来自同一不可变声明快照。
- 将 JSON 快照原子替换写入磁盘，并将审计事件只追加写入 JSONL 日志。
- 提供可筛选、稳定分页且带完整性状态的事件时间线，以及流程进度、工作台汇总、健康检查和版本化规则目录接口。

## 环境要求

需要 Go 1.22 或更高版本。项目仅使用 Go 标准库。

## 构建

在项目根目录执行：

```bash
go build ./cmd/server
```

## 运行

默认只监听高位回环地址 `127.0.0.1:19081`，数据写入 `data/`：

```bash
go run ./cmd/server
```

可显式指定监听地址和数据目录：

```bash
go run ./cmd/server -addr=127.0.0.1:19281 -data=./data
```

未显式传入 `-addr` 时，也可通过 `PORT` 指定端口，服务会绑定 `127.0.0.1:<PORT>`。出于本地数据保护考虑，`-addr` 只接受回环主机。启动后访问 `http://127.0.0.1:19081/` 使用浏览器工作台。

## 测试与自检

运行全部单元和集成测试：

```bash
go test ./...
```

运行有界真实 HTTP 自检。该命令使用临时数据目录，在指定回环地址启动服务，经公开 API 完成建档、制造四类问题、整改、复核、批准、发布和声明导出后自动关闭：

```bash
go run ./cmd/server -selftest -addr=127.0.0.1:19081
```

## 数据文件

`-data` 指定目录中包含：

- `cases/<case_id>.json`：验收案、最新内容及不可变内容修订历史的 JSON 快照。
- `events.jsonl`：只追加的不可变事件时间线。
- `idempotency.json`：`request_id` 对应的既有操作结果。

应用启动时会校验快照标识、内容修订、事件引用和事件修订顺序。请在服务停止后备份或迁移这些文件，不要在运行时手工修改。

## HTTP 接口

`GET /api/cases` 接受 `q`、`owner`、`status`、`date_from`、`date_to`、`sort`、`order`、`cursor` 和 `limit` 查询参数。`sort` 支持 `target_publish_date` 与 `updated_at`。

内容历史接口包括 `GET /api/cases/{id}/content/revisions`、`GET /api/cases/{id}/content/diff?from=<revision>&to=<revision>` 和 `POST /api/cases/{id}/content/restore`。复审继续使用 `POST /api/cases/{id}/audit`。

批量整改和复核分别使用 `POST /api/cases/{id}/evidence/batch` 与 `POST /api/cases/{id}/reviews/batch`。声明预览使用 `GET /api/cases/{id}/declaration/preview?approver=<name>`，批准、发布、HTML 页面和 JSON 导出沿用现有接口。时间线筛选使用 `GET /api/cases/{id}/timeline`，支持 `event_type`、`actor`、`revision_from`、`revision_to`、`cursor` 与 `limit`。

所有写请求使用 `application/json`，必须携带当前 `revision`、操作人和唯一 `request_id`；也可用 `Idempotency-Key` 请求头提供请求标识。
