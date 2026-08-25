# BENZHI_README

## 项目说明
- 项目：benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3
- 项目用途：已完整实现面向公共服务内容团队的无障碍文档整改验收台，提供建档、结构化内容修订、确定性规则审查、版本化整改证据、人工复核、批准发布、可打印声明、JSON 导出、乐观并发、请求幂等和不可变事件时间线。
- Go 工具链：`golang:1.22`
- 前端工具链：原生 HTML、CSS 和 JavaScript，由 Go 服务直接提供

## 项目描述
- 项目名称：无障碍文档整改验收台
- 项目概述：面向公共服务内容团队的无障碍文档整改验收应用，将一份待发布文档从建档、结构审查、整改复核推进到合格声明发布，并保留可追溯证据。项目按 standard 档位规划不少于 2000 行真实生产 Go 代码、至少 24 个生产 Go 文件，根目录提供简体中文 README.md，说明用途以及标准构建、运行和测试方式。
- 核心工作流：编辑创建文档验收案并录入结构化内容，系统执行无障碍规则审查并生成问题清单，编辑逐项提交整改证据，审核员复核全部问题后批准，系统发布合格声明并将验收案关闭。
- 对外接口：由 Go HTTP 服务提供原生 HTML、CSS 和 JavaScript 的浏览器工作台，包含验收案列表、结构化文档编辑区、规则问题与整改证据区、复核操作区及事件时间线；不引入 Node 构建链。服务支持 -addr=127.0.0.1:<port>，也支持读取 PORT 并绑定 127.0.0.1:<PORT>，默认监听 127.0.0.1:19081，禁止默认绑定 8080、80、3000 或 0.0.0.0。

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/server -selftest -addr=127.0.0.1:19081
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3-arm64 linux/arm64
docker run -it benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -selftest -addr=127.0.0.1:19081`
