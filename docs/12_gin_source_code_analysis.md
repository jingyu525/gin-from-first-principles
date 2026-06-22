# 第十二章：Gin 源码解剖

## 1. gin.go 逐行解析

### 1.1 Engine 结构体
```go
type Engine struct {
    RouterGroup  // 嵌入 RouterGroup，支持 Group 路由
    
    // 核心技术：基数树森林
    trees methodTrees  // 每种 HTTP 方法对应一棵基数树
    
    // 性能优化：对象池
    pool sync.Pool
    
    // 配置
    RedirectTrailingSlash bool
    RedirectFixedPath     bool
    HandleMethodNotAllowed bool
    ForwardedByClientIP   bool
    
    // 其他
    HTMLRender            render.HTMLRender
    FuncMap               template.FuncMap
    trustedProxies       []string
    maxMultipartMemory   int64
}
```

### 1.2 New() 函数
```go
func New() *Engine {
    engine := &Engine{
        RouterGroup: RouterGroup{
            Handlers: nil,
            basePath: "/",
            root:     true,
        },
        trees: make(methodTrees, 0, 9),
    }
    
    // 关键：初始化对象池
    engine.pool.New = func() interface{} {
        return engine.allocateContext()
    }
    
    return engine
}
```

### 1.3 ServeHTTP 方法（核心）
```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // 1. 从对象池获取 Context
    c := engine.pool.Get().(*Context)
    c.writermem.reset()
    c.Request = req
    c.Writer = &c.writermem
    
    // 2. 查找路由
    engine.handleHTTPRequest(c)
    
    // 3. 执行完后归还到对象池
    engine.pool.Put(c)
}
```

---

## 2. context.go 逐行解析

### 2.1 Context 结构体
```go
type Context struct {
    writermem responseWriter
    Request   *http.Request
    Writer    ResponseWriter
    
    Params   Params  // 路由参数
    handlers HandlersChain  // 中间件链
    index    int8    // 当前执行到哪个中间件（注意是 int8，不是 int）
    
    // Keys 用于中间件间传递数据
    Keys map[string]interface{}
    
    // 错误处理
    errors errorMsgs
    
    // 其他
    Accepted []string
    query   url.Values
    postForm url.Values
}
```

### 2.2 Next() 方法
```go
func (c *Context) Next() {
    c.index++  // 移动到下一个中间件
    for c.index < int8(len(c.handlers)) {
        c.handlers[c.index](c)
        c.index++
    }
}
```
**注意**：Gin 的 `Next()` 实现与我们的 Mini-Gin 略有不同，但核心思想一致。

### 2.3 Abort() 方法
```go
func (c *Context) Abort() {
    c.index = abortIndex  // abortIndex = 63（int8 的最大值）
}
```
**技巧**：利用 `int8` 的最大值（63）作为终止索引。

---

## 3. recovery 中间件

### 3.1 如何利用 panic 和 recover 防止进程崩溃
```go
func CustomRecovery() HandlerFunc {
    return func(c *Context) {
        defer func() {
            if err := recover(); err != nil {
                // 打印堆栈
                debug.PrintStack()
                
                // 返回 500
                c.AbortWithStatus(500)
            }
        }()
        
        c.Next()
    }
}
```

### 3.2 为什么 recover 必须直接在 defer 中？
**原因**：`recover()` 只有在 **直接调用** 的 `defer` 函数中才有效。
```go
// ❌ 错误：recover 在嵌套函数中，无效
defer func() {
    if err := recover(); err != nil {
        handleError(err)  // 在另一个函数中调用 recover
    }
}()

// ✅ 正确：recover 直接在 defer 中
defer func() {
    if err := recover(); err != nil {
        // 直接处理
        c.AbortWithStatus(500)
    }
}()
```

---

## 4. 性能优化技巧

### 4.1 使用 sync.Pool 减少内存分配
**对比**：
```go
// ❌ 每次请求都创建新的 Context
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    c := &Context{}  // 新的内存分配
    // ...
}

// ✅ 从对象池复用 Context
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    c := engine.pool.Get().(*Context)  // 复用旧对象
    defer engine.pool.Put(c)  // 用完归还
    // ...
}
```

### 4.2 使用 []byte 池减少字符串分配
Gin 使用 `bytes.Buffer` 池来优化响应写入。

---

## 5. 下一章预告

Gin 源码解剖完成！最后我们将学习 **分布式下的 Gin 避坑指南**：
- Context 的生命周期陷阱
- TraceID 的传播
- 超时传递的一致性

**准备成为 Gin 专家！** 🏆
