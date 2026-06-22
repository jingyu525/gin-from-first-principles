# 第七章：Router Group 的推导

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
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
    return &RouterGroup{
        prefix:     g.prefix + prefix,
        middleware:  g.middleware,
        parent:      g,
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
    // 注册路由时，包含所有继承的中间件
    e.router.addRoute("GET", fullPath, allHandlers...)
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
