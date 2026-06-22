# 第四章：Gin 核心组件推导

## 1. Context 结构体设计

### 1.1 封装 Req/Res + KV 存储
Gin 的 `Context` 是整个框架的核心，它封装了：
- **输入**：`*http.Request`
- **输出**：`http.ResponseWriter`
- **存储**：中间件共享的 KV 数据
- **索引**：控制中间件执行流程

```go
type Context struct {
    Request *http.Request
    Writer  http.ResponseWriter
    
    // KV 存储
    Keys map[string]interface{}
    
    // 中间件控制
    handlers []Handler  // 中间件链
    index    int        // 当前执行到哪个中间件
}
```

### 1.2 设置/获取 KV 的方法
```go
func (c *Context) Set(key string, value interface{}) {
    c.Keys[key] = value
}

func (c *Context) Get(key string) interface{} {
    return c.Keys[key]
}
```

---

## 2. Radix Tree（基数树）

### 2.1 为什么 Map 不行？
如果使用 `map[string]Handler`，无法处理层级关系：
```go
// 无法区分这两个路由
"/users/:id"    -> handler1
"/users/profile" -> handler2
```

### 2.2 基数树的节点结构
```go
type node struct {
    path     string            // 当前节点路径
    children map[string]*node  // 子节点
    handler  Handler           // 处理函数
    isParam  bool              // 是否是参数节点（如 :id）
}
```

### 2.3 动态路由的参数提取逻辑
当请求 `/users/123` 时：
1.  从根节点开始匹配
2.  遇到 `:id` 节点，提取 `123`
3.  将参数存入 `Context.Params`

```go
func (c *Context) Param(key string) string {
    return c.Params[key]
}
```

---

## 3. 责任链模式（Chain）

### 3.1 Next() 函数的推导
**核心思想**：每个中间件可以决定是否调用下一个中间件。

```go
func (c *Context) Next() {
    c.index++  // 移动到下一个中间件
    if c.index < len(c.handlers) {
        c.handlers[c.index](c)
    }
}
```

### 3.2 中间件的写法
```go
func Logger() Handler {
    return func(c *Context) {
        start := time.Now()
        c.Next()  // 执行后续中间件
        log.Printf("Duration: %v", time.Since(start))
    }
}
```

### 3.3 执行流程图
```
请求到来
  ↓
Logger 中间件 (before)
  ↓
Auth 中间件 (before)
  ↓
实际 Handler
  ↓
Auth 中间件 (after)
  ↓
Logger 中间件 (after)
  ↓
响应返回
```

---

## 4. 下一章预告

理论推导完成！接下来我们将 **手写 Mini-Gin**，实现以上所有组件：
1.  Context 实现
2.  Radix Tree 实现
3.  Engine 串联一切

**准备动手写代码！** 💻
