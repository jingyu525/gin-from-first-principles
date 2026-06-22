# 第五章：实现 Mini-Gin（约 200 行）

## 1. 环境准备：Go Module 初始化

> **对应代码**：`mini-gin/05-basic/`

### 1.1 创建项目
```bash
mkdir mini-gin
cd mini-gin
go mod init github.com/jingyu525/mini-gin/05-basic
```

### 1.2 项目结构
```
05-basic/
├── go.mod
├── context.go   # Context 定义
├── router.go    # Radix Tree 路由 + Engine
└── main.go      # 启动入口
```

---

## 2. Context 实现：请求容器与 Next() 逻辑

### 2.1 定义 Context 结构体
```go
type Context struct {
    Writer   http.ResponseWriter
    Request  *http.Request
    Params   map[string]string  // 路由参数
    index    int                // 中间件索引
    handlers []Handler          // 中间件链
}
```

### 2.2 实现 Next() 方法
```go
func (c *Context) Next() {
    c.index++
    if c.index < len(c.handlers) {
        c.handlers[c.index](c)
    }
}
```

### 2.3 实现 Abort() 方法
```go
func (c *Context) Abort() {
    c.index = len(c.handlers)
}
```

---

## 3. Router 实现：简易 Radix Tree

### 3.1 定义节点结构
```go
type node struct {
    path     string
    children map[string]*node
    handler  Handler
    isParam  bool
}
```

### 3.2 添加路由（插入节点）
```go
func (r *Router) addRoute(method, path string, handler Handler) {
    parts := parsePath(path)  // 将路径分割成部分
    root := r.root[method]
    
    for _, part := range parts {
        child := root.children[part]
        if child == nil {
            child = &node{
                path:    part,
                isParam: part[0] == ':',
            }
            root.children[part] = child
        }
        root = child
    }
    root.handler = handler
}
```

### 3.3 查找路由（匹配节点）
```go
func (r *Router) getRoute(method, path string) (*node, map[string]string) {
    params := make(map[string]string)
    parts := parsePath(path)
    root := r.root[method]
    
    for i, part := range parts {
        found := false
        for _, child := range root.children {
            if child.isParam {
                params[child.path[1:]] = part
                found = true
            } else if child.path == part {
                found = true
            }
            if found {
                root = child
                break
            }
        }
        if !found {
            return nil, nil
        }
    }
    return root, params
}
```

---

## 4. Engine 实现：接入 net/http

### 4.1 定义 Engine 结构体
```go
type Engine struct {
    router   *Router
    handlers []Handler  // 全局中间件
}
```

### 4.2 实现 ServeHTTP 接口
```go
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 查找路由
    node, params := e.router.getRoute(r.Method, r.URL.Path)
    if node == nil || node.handler == nil {
        w.WriteHeader(404)
        return
    }
    
    // 创建 Context
    c := &Context{
        Writer:   w,
        Request:  r,
        Params:   params,
        index:    -1,
        handlers: append(e.handlers, node.handler),
    }
    
    // 执行中间件链
    c.Next()
}
```

### 4.3 启动服务器
```go
func (e *Engine) Run(addr string) error {
    return http.ListenAndServe(addr, e)
}
```

---

## 5. 运行与测试

### 5.1 编写 main.go
```go
func main() {
    r := New()

    r.Use(func(c *Context) {
        fmt.Println("[Log] Request:", c.Request.Method, c.Request.URL.Path)
        c.Next()
    })

    r.GET("/users/:id", func(c *Context) {
        id := c.Param("id")
        c.JSON(200, map[string]string{"id": id})
    })

    r.Run(":8080")
}
```

### 5.2 启动服务
```bash
cd mini-gin/05-basic && go run .
```

### 5.3 发送请求
```bash
curl http://localhost:8080/users/123
# 输出：{"id":"123"}
```

---

## 6. 总结

我们已经实现了一个 **简易版的 Gin**：
- ✅ Context 封装
- ✅ Radix Tree 路由
- ✅ 中间件责任链
- ✅ 参数提取

**下一章**：我们将优化性能，引入 `sync.Pool` 减少 GC 压力。 🚀
