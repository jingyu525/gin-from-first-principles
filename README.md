# Gin From First Principles
### 从字节流推导至分布式网关

[![Go Report Card](https://goreportcard.com/badge/github.com/jingyu525/gin-from-first-principles)](https://goreportcard.com/report/github.com/jingyu525/gin-from-first-principles)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **"我不只是教你用 Gin，我是带你重新发明 Gin。"**

## 📖 项目简介

市面上的教程大多告诉你 **How**（怎么用），但很少解释 **Why**（为什么是这样）。

本项目试图回归 **First Principles（第一性原理）**，从最底层的网络字节流开始，一步步推演出现代 Go Web 开发的全貌：

1.  **无抽象**：从 `net` 包手写 TCP 服务器，推导 HTTP 协议的必要性。
2.  **造轮子**：实现 Mini-Gin，包含 Radix Tree 路由与 Context 池化。
3.  **架构升维**：从单体 Gin 演进到 API Gateway 与 Service Mesh。

## 🗺️ 全链路推导路线图

```
字节流 (Bytes)
    ↓
TCP 连接 (net.Listen)
    ↓
HTTP 协议解析 (net/http)
    ↓
路由抽象 (Map -> Radix Tree)
    ↓
上下文管理 (Context & sync.Pool)
    ↓
中间件责任链 (Chain & Abort)
    ↓
微服务拆分 (Service Layer)
    ↓
流量入口 (API Gateway)
    ↓
服务治理 (Service Mesh)
```

## 🚀 快速开始

### 运行 Mini-Gin (Core)
```bash
cd mini-gin
go run main.go
curl http://localhost:8080/users/123
```

### 运行完整微服务架构 (Arch)

#### 方式一：使用 Docker Compose（推荐）
```bash
# 一键启动所有服务
docker-compose up --build

# 测试
curl http://localhost:8080/users/123
curl -H "Authorization: Bearer your-token" http://localhost:8080/orders/456
```

#### 方式二：本地运行
```bash
# 终端 1：启动用户服务
cd arch/user-service && go run main.go

# 终端 2：启动订单服务
cd arch/order-service && go run main.go

# 终端 3：启动 API Gateway
cd arch/gateway && go run main.go
```

## 📚 教程目录

本项目包含一套完整的从零开始学习 Go Web 开发的教程，位于 `docs/` 目录：

| 章节 | 标题 | 文件 |
|------|------|------|
| 01 | 剥离抽象 —— 从 Socket 到 HTTP | [01_from_socket_to_http.md](docs/01_from_socket_to_http.md) |
| 02 | net/http 的诞生 | [02_the_birth_of_net_http.md](docs/02_the_birth_of_net_http.md) |
| 03 | 诊断原生缺陷 | [03_diagnosing_native_defects.md](docs/03_diagnosing_native_defects.md) |
| 04 | Gin 核心组件 | [04_gin_core_components.md](docs/04_gin_core_components.md) |
| 05 | Mini-Gin 实现 | [05_mini_gin_implementation.md](docs/05_mini_gin_implementation.md) |
| 06 | 性能优化：Pool vs GC | [06_performance_pool_vs_gc.md](docs/06_performance_pool_vs_gc.md) |
| 07 | Router Group 设计 | [07_router_group_design.md](docs/07_router_group_design.md) |
| 08 | 控制流设计 | [08_control_flow_design.md](docs/08_control_flow_design.md) |
| 09 | 分布式视角 | [09_distributed_perspective.md](docs/09_distributed_perspective.md) |
| 10 | API Gateway 实现 | [10_api_gateway_implementation.md](docs/10_api_gateway_implementation.md) |
| 11 | Service Mesh 解密 | [11_service_mesh_demystified.md](docs/11_service_mesh_demystified.md) |
| 12 | Gin 源码分析 | [12_gin_source_code_analysis.md](docs/12_gin_source_code_analysis.md) |
| 13 | 避坑指南（分布式下的 Gin） | [13_pitfall_guide.md](docs/13_pitfall_guide.md) |

## 📂 项目结构

```
.
├── docs/            # 教程文档（13 章完整教程）
├── mini-gin/        # 核心：从零实现的 Web 框架
├── arch/            # 架构：Gateway & Service 示例
└── README.md
```

## 🎯 适合谁阅读？

- 厌倦了 CRUD，想深入理解框架底层原理的 **Gopher**。
- 想从"写接口"进阶到"系统设计"的 **后端工程师**。
- 对 **高性能**、**高并发** 实现机制感兴趣的开发者。

## 📜 License

MIT
