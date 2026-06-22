# 第八章：控制流设计（Abort & Error）

## 1. Abort 机制

### 1.1 推导：如何在中间件中强制终止后续执行？
**场景**：鉴权失败，不想执行后续中间件和 Handler。

#### 错误方案：直接 return
```go
func AuthMiddleware(c *Context) {
    if c.Request.Header.Get("Token") == "" {
        c.Writer.WriteHeader(401)
        return  // 只能终止当前中间件，后续中间件仍会执行
    }
    c.Next()
}
```

#### 正确方案：c.Abort()
```go
func (c *Context) Abort() {
    c.index = len(c.handlers)  // 将索引跳到末尾
}
```

### 1.2 实现：c.Abort() 与索引越界技巧
当 `c.index == len(c.handlers)` 时，`c.Next()` 不会执行任何处理函数：
```go
func (c *Context) Next() {
    c.index++
    if c.index < len(c.handlers) {  // 关键判断
        c.handlers[c.index](c)
    }
}
```

---

## 2. 错误处理链

### 2.1 推导：Error 作为一种特殊的中间件
**场景**：统一处理 panic 和运行时错误。

#### 实现思路
```go
func RecoveryMiddleware() Handler {
    return func(c *Context) {
        defer func() {
            if err := recover(); err != nil {
                c.Writer.WriteHeader(500)
                fmt.Fprintf(c.Writer, "Internal Server Error")
            }
        }()
        c.Next()
    }
}
```

### 2.2 统一的异常捕获与返回
```go
func (c *Context) Fail(code int, msg string) {
    c.index = len(c.handlers)  // 终止后续执行
    c.Writer.WriteHeader(code)
    fmt.Fprintf(c.Writer, msg)
}

// 在 Handler 中使用
func Handler(c *Context) {
    if err := someFunc(); err != nil {
        c.Fail(500, "Internal Error")
        return
    }
    c.JSON(200, map[string]string{"status": "ok"})
}
```

---

## 3. 完整的执行流程图

```
请求到来
  ↓
RecoveryMiddleware (defer recover)
  ↓
LoggerMiddleware (before)
  ↓
AuthMiddleware
  ├─ 鉴权成功 → c.Next() → Handler → c.Abort()
  └─ 鉴权失败 → c.Abort() → 直接返回 401
  ↓
LoggerMiddleware (after)
  ↓
RecoveryMiddleware (检查 recover)
  ↓
响应返回
```

---

## 4. 下一章预告

控制流设计完成！接下来我们将进入 **分布式视角**：
- 为什么一个 Gin 不够？
- 如何用 Gin 实现 API Gateway？

**准备迎接架构升级！** 🚀
