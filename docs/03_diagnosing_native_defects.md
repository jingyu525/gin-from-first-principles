# 第三章：诊断原生缺陷

## 1. 上下文（Context）的缺失

### 1.1 痛点：数据无法在中间件间传递
在 `net/http` 中，如果我们有两个中间件：
- 中间件 A：解析 JWT，得到 `user_id`
- 中间件 B：记录日志，需要 `user_id`

**问题**：如何让 B 拿到 A 的数据？

#### 错误方案：全局变量
```go
var GlobalUserID string  // 不安全！并发请求会互相覆盖
```

#### 正确推导：请求级别的"储物柜"
我们需要一个结构体，生命周期与请求相同：
```go
type Context struct {
    Request  *http.Request
    Writer   http.ResponseWriter
    storage  map[string]interface{}  // 中间件共享数据
}
```

---

## 2. 路由的孱弱

### 2.1 痛点：DefaultServeMux 只能前缀匹配
```go
// 想要匹配 /users/123
// 但 DefaultServeMux 只能注册固定路径
http.HandleFunc("/users/123", handler)  // 只能匹配这一个
```

### 2.2 推导：如何实现 /users/:id？
我们需要 **动态路由**，支持参数提取：
```
路径模式：/users/:id
实际请求：/users/123
提取结果：id = "123"
```

**技术方案**：基数树（Radix Tree）
- 将路径按 `/` 分割成节点
- `:id` 是一个特殊节点，能匹配任意字符串

---

## 3. 中间件嵌套地狱

### 3.1 痛点：为了日志和鉴权，函数层层包裹
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    // 日志
    log.Println("Request start")
    defer log.Println("Request end")
    
    // 鉴权
    if r.Header.Get("Token") == "" {
        w.WriteHeader(401)
        return
    }
    
    // 实际处理
    w.Write([]byte("Hello"))
}
```

**问题**：每个 Handler 都要写一遍日志和鉴权代码。

### 3.2 推导：如何扁平化管理执行流程？
我们需要一个 **责任链模式**：
```go
func LoggerMiddleware(next Handler) Handler {
    return func(c *Context) {
        log.Println("Start")
        next(c)
        log.Println("End")
    }
}

func AuthMiddleware(next Handler) Handler {
    return func(c *Context) {
        if c.Request.Header.Get("Token") == "" {
            c.Writer.WriteHeader(401)
            return
        }
        next(c)
    }
}

// 使用
r.GET("/users", LoggerMiddleware(AuthMiddleware(Handler)))
```

**更优雅的方案**：Gin 的 `c.Next()` 机制（下一章详解）。

---

## 4. 下一章预告

我们将基于以上三个痛点，推导 **Gin 的核心组件**：
1.  Context 结构体设计
2.  Radix Tree 路由实现
3.  责任链模式（Chain）

**准备进入 Gin 的世界！** 🚀
