package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// Handler 处理函数类型
type Handler func(*Context)

// Context 封装了请求上下文，提供便捷的方法
type Context struct {
    Writer  http.ResponseWriter
    Request *http.Request
    Params  map[string]string // 路由参数，如 :id
    index   int               // 中间件索引
    handlers []Handler        // 中间件和处理函数链
}

// Next 执行下一个中间件
func (c *Context) Next() {
    c.index++
    if c.index < len(c.handlers) {
        c.handlers[c.index](c)
    }
}

// Abort 终止中间件链
func (c *Context) Abort() {
    c.index = len(c.handlers)
}

// JSON 返回 JSON 响应
func (c *Context) JSON(code int, obj interface{}) {
    c.Writer.Header().Set("Content-Type", "application/json")
    c.Writer.WriteHeader(code)
    encoder := json.NewEncoder(c.Writer)
    if err := encoder.Encode(obj); err != nil {
        c.Writer.WriteHeader(http.StatusInternalServerError)
        c.Writer.Write([]byte(`{"error": "internal server error"}`))
    }
}

// String 返回纯文本响应
func (c *Context) String(code int, format string, values ...interface{}) {
    c.Writer.Header().Set("Content-Type", "text/plain")
    c.Writer.WriteHeader(code)
    fmt.Fprintf(c.Writer, format, values...)
}

// Param 获取路由参数
func (c *Context) Param(key string) string {
    if c.Params == nil {
        return ""
    }
    return c.Params[key]
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
    return c.Request.URL.Query().Get(key)
}

// reset 重置 Context 状态，用于对象池复用
func (c *Context) reset() {
    c.Writer = nil
    c.Request = nil
    c.Params = nil
    c.index = -1
    c.handlers = nil
}
