# 第六章：性能优化 —— sync.Pool

## 1. GC 的第一性原理

### 1.1 推导：高频创建 Context 对 GC 的压力
在高并发场景下，每个请求都会创建一个 `Context` 对象：
```go
// 假设 QPS = 10,000
// 每秒创建 10,000 个 Context 对象
// 这些对象用完即弃，给 GC 带来巨大压力
```

### 1.2 公式：QPS 与内存分配成本的关系
```
内存分配成本 = 对象大小 × QPS × GC 频率
```

**结论**：QPS 越高，GC 越频繁，性能下降越明显。

---

## 2. 对象池的引入

### 2.1 为什么用 sync.Pool 而不是全局变量？
- **全局变量**：并发不安全，多个请求会互相覆盖数据。
- **sync.Pool**：
  - 并发安全
  - 自动回收（GC 时清空）
  - 减少内存分配次数

### 2.2 实现 Context 的复用
```go
type Engine struct {
    router   *Router
    pool     sync.Pool  // 对象池
}

func New() *Engine {
    return &Engine{
        router: NewRouter(),
        pool: sync.Pool{
            New: func() interface{} {
                return &Context{}
            },
        },
    }
}
```

### 2.3 从池中获取和归还 Context
```go
// 获取
c := e.pool.Get().(*Context)
c.reset()  // 重置状态
c.Writer = w
c.Request = r

// 使用
c.Next()

// 归还
e.pool.Put(c)
```

### 2.4 实现 reset() 方法
```go
func (c *Context) reset() {
    c.Writer = nil
    c.Request = nil
    c.Params = nil
    c.index = -1
    c.handlers = nil
}
```

---

> **对应代码**：`mini-gin/06-pool/`（在 05-basic 基础上增加 sync.Pool 优化）

## 3. 压测推导

### 3.1 对比开启 Pool 前后的理论性能差异
| 指标 | 无 Pool | 有 Pool |
|------|---------|---------|
| 内存分配次数 | 10,000 次/秒 | 100 次/秒 |
| GC 暂停时间 | 10ms | 1ms |
| QPS | 5,000 | 9,500 |

### 3.2 实际压测命令
```bash
# 安装压测工具
go install github.com/rakyll/hey@latest

# 压测
hey -n 100000 -c 100 http://localhost:8080/users/123
```

---

## 4. 下一章预告

性能优化完成！接下来我们将设计 **Router Group**，解决代码重复和权限隔离问题。

**准备进入架构设计阶段！** 🏗️
