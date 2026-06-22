# 第七章：Router Group 的推导

> **对应代码**：`mini-gin/07-groups/`（在 06-pool 基础上增加 RouterGroup）

## 1. 问题：代码重复与权限隔离

### 1.1 场景描述
假设我们有两组 API：
- `/v1/users` 和 `/v1/orders`（旧版本）
- `/v2/users` 和 `/v2/orders`（新版本）

**痛点**：每个路由都要写一遍前缀 `/v1` 或 `/v2`。

### 1.2 权限隔离的需求
- `/v1/*` 需要旧版鉴权中间件
- `/v2/*` 需要新版 JWT 鉴权中间件

**问题**：如何避免在每个路由上重复写中间件？

---

## 2. 推导 Group 结构

### 2.1 前缀（Prefix）的拼接逻辑
```go
type RouterGroup struct {
    prefix      string
    middleware  []Handler
    parent      *RouterGroup  // 支持嵌套
    engine      *Engine       // 指向主引擎，用于注册路由
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
    return &RouterGroup{
        prefix:     g.prefix + prefix,
        middleware:  append([]Handler{}, g.middleware...),  // 复制切片，避免污染父 Group
        parent:      g,
        engine:      g.engine,  // 传递 Engine 引用
    }
}
```

### 2.2 中间件继承：父 Group 影响子 Group
```go
func (g *RouterGroup) Use(middleware ...Handler) {
    g.middleware = append(g.middleware, middleware...)
}

func (g *RouterGroup) GET(path string, handler Handler) {
    fullPath := g.prefix + path
    allHandlers := append(g.middleware, handler)
    // 通过闭包将中间件链注入 Context，然后注册到 Engine 的路由器
    g.engine.router.addRoute("GET", fullPath, func(c *Context) {
        c.handlers = allHandlers
        c.Next()
    })
}
```

---

## 3. 实际使用 example

### 3.1 定义两个版本的 API
```go
func main() {
    r := New()
    
    // V1 组
    v1 := r.Group("/v1")
    v1.Use(OldAuthMiddleware())
    v1.GET("/users", handler1)
    v1.GET("/orders", handler2)
    
    // V2 组
    v2 := r.Group("/v2")
    v2.Use(JWTAuthMiddleware())
    v2.GET("/users", handler3)
    v2.GET("/orders", handler4)
    
    r.Run(":8080")
}
```

### 3.2 执行流程
```
请求 /v1/users
  ↓
匹配到 v1 组
  ↓
执行 OldAuthMiddleware
  ↓
执行 handler1
```

---

## 4. 下一章预告

Router Group 设计完成！接下来我们将设计 **控制流**：
- `c.Abort()` 的实现
- 错误处理链

**准备深入中间件机制！** 🔧
