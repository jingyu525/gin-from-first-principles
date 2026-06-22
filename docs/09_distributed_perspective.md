# 第九章：单体的局限

## 1. 推导拆分：为什么一个 Gin 不够？

### 1.1 发布风险
**场景**：一个单体应用包含用户、订单、支付等功能。
- **问题**：修改支付模块，可能导致用户模块崩溃。
- **推导**：应该将不同业务拆分成独立服务。

### 1.2 数据库连接数
**场景**：单体应用有 100 个实例，每个实例占用 10 个数据库连接。
- **计算**：100 × 10 = 1000 个连接。
- **推导**：拆分后，每个服务只需少量连接。

### 1.3 团队协作
**场景**：10 个团队共用一个代码仓库。
- **问题**：代码冲突频繁，发布互相阻塞。
- **推导**：每个团队维护自己的服务。

---

## 2. Service 层定义

### 2.1 什么是 Service？
**定义**：一个独立的、可部署的 Gin 实例。
```go
// user-service/main.go
func main() {
    r := gin.Default()
    r.GET("/users/:id", handleGetUser)
    r.Run(":8081")
}

// order-service/main.go
func main() {
    r := gin.Default()
    r.GET("/orders/:id", handleGetOrder)
    r.Run(":8082")
}
```

### 2.2 服务间如何通信？
#### 方案 A：HTTP（简单，推荐）
```go
// 在 gateway 中调用 user-service
resp, err := http.Get("http://localhost:8081/users/" + userID)
```

#### 方案 B：RPC（高性能，复杂）
- 使用 gRPC 或 Thrift
- 需要定义 Proto 文件
- 适合内部服务调用

---

## 3. 服务拆分的原则

### 3.1 单一职责原则（SRP）
每个服务只做一件事：
- User Service：管理用户
- Order Service：管理订单
- Payment Service：处理支付

### 3.2 数据库拆分
**不要共享数据库！**
```
❌ 错误：所有服务连同一个数据库
✅ 正确：每个服务有自己的数据库
```

---

## 4. 下一章预告

单体拆分完成！接下来我们将实现 **API Gateway**：
- 网关的第一性原理
- 用 Gin 实现反向代理

**准备进入分布式核心！** 🌐
