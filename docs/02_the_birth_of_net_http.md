# 第二章：net/http 的诞生

## 1. 抽象的必要性

在第一章中，我们手写 TCP 服务器，手动解析 HTTP 协议头。这种方式极其脆弱，且每个程序员都要重复造轮子。

### 1.1 将"字节解析"封装为 Request 结构体
我们需要一个统一的方式来处理 HTTP 请求：
```go
type Request struct {
    Method string
    Path   string
    Header map[string]string
    Body   []byte
}
```

### 1.2 将"写入响应"封装为 ResponseWriter
我们需要一个统一的方式来构造 HTTP 响应：
```go
type ResponseWriter interface {
    Header() http.Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
```

---

## 2. 路由映射的推导

### 2.1 为什么需要 HandleFunc？
当请求到达时，我们需要根据 `Method` 和 `Path` 找到对应的处理函数。这本质上是一个 **Key-Value 映射**：
```
Key = "GET:/hello"
Value = func(w http.ResponseWriter, r *http.Request)
```

### 2.2 DefaultServeMux 的实现原理
Go 的 `net/http` 包提供了一个默认的路由器：
```go
var DefaultServeMux = &ServeMux{}

type ServeMux struct {
    mu    sync.RWMutex
    m     map[string]muxEntry
}
```

---

## 3. 最简 HTTP 服务器的全链路推演

当我们写下这段代码时：
```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello"))
})
http.ListenAndServe(":8080", nil)
```

实际上发生了以下 **6 个步骤**：

1.  **Kernel**：内核接受 TCP 握手（三次握手）。
2.  **net/http**：`http.Server` 接管连接。
3.  **Parse**：读取字节流，解析出 `Method=GET`, `Path=/hello`, `Header`。
4.  **Route**：查询 `DefaultServeMux`，找到对应的 `Handler`。
5.  **Execute**：调用你的处理函数。
6.  **Write**：将返回值通过 TCP 写回客户端。

---

## 4. 遗留的痛点

虽然 `net/http` 解决了协议解析问题，但我们留下了 **三个隐患**，这将是我们下一章推导 **Gin** 的起点：

1.  **Context 缺失**：如何在中间件间传递变量？
2.  **路由孱弱**：如何处理 `/users/:id`？
3.  **中间件混乱**：如何优雅地组织日志、鉴权？

**下一章预告**：我们将诊断 `net/http` 的缺陷，并推导出 Gin 的核心数据结构。
