# 第一章：剥离抽象 —— 从 Socket 到 HTTP

## 1. 第一性原理假设

在开始写代码之前，我们必须抛弃所有关于 "Web 框架" 的常识。
我们只承认以下事实：

1.  **物理限制**：计算机只能通过网卡收发 **字节流 (Byte Stream)**。
2.  **协议约定**：为了理解这些字节，双方必须遵循 **HTTP 文本协议**。
3.  **OS 限制**：应用无法直接操作网卡，必须通过 **系统调用 (Syscall)**。

基于这些事实，我们推导：任何 Web 服务，本质上都是一个 **"读取字节 -> 解析文本 -> 生成字节"** 的程序。

---

## 2. 推导：为什么需要 TCP？

如果我们直接用 Go 的 `net` 包监听端口，会发生什么？

### 2.1 监听（Listen）
操作系统需要知道谁来处理 8080 端口的流量。
```go
listener, _ := net.Listen("tcp", ":8080")
```

### 2.2 连接（Accept）
TCP 连接是一个持续的数据流。
```go
conn, _ := listener.Accept()
```

### 2.3 并发（Goroutine）
如果我们串行处理连接，第二个请求必须等待第一个完成。
**推导结论**：必须为每个连接启动一个独立的执行单元（Goroutine）。

```go
for {
    conn, _ := listener.Accept()
    go handleConn(conn) // 并发处理
}
```

---

## 3. 推导：HTTP 协议的必要性

现在，我们拿到了 `conn`，里面只有字节。

### 3.1 原始字节的困境
假设客户端发送：

```
GET /hello HTTP/1.1\r\nHost: localhost\r\n\r\n
```

我们的程序需要手动切割字符串，判断 Method 和 Path。这极其脆弱。

### 3.2 抽象诞生
为了不让每个程序员都去解析字符串，我们需要一层封装：
- **Request**：把字节流转换成结构体。
- **ResponseWriter**：帮我们拼装 HTTP 响应头。

这就是 `net/http` 包存在的根本原因。

---

## 4. 最简服务器的全链路推演

当我们写下这段"最简单"的代码时：

```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
})
http.ListenAndServe(":8080", nil)
```

实际上发生了：

1.  **Kernel**：内核接受 TCP 握手。
2.  **net/http**：Server 接管连接。
3.  **Parse**：读取字节，解析出 `Method=GET`, `Path=/hello`。
4.  **Route**：查询内部的 Map（DefaultServeMux），找到对应的 Handler。
5.  **Execute**：调用你的函数。
6.  **Write**：将返回值通过 TCP 写回客户端。

---

## 5. 遗留的痛点

虽然 `net/http` 解决了协议问题，但我们留下了三个隐患，这将是我们下一章推导 **Gin** 的起点：

1.  **Context 缺失**：如何在中间件间传递变量？
2.  **路由孱弱**：如何处理 `/users/:id`？
3.  **中间件混乱**：如何优雅地组织日志、鉴权？

**下一章预告**：我们将诊断 `net/http` 的缺陷，并推导出 Gin 的核心数据结构。
