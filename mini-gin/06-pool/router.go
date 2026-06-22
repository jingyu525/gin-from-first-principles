package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// node 基数树节点
type node struct {
	path     string           // 当前节点路径
	children map[string]*node // 子节点
	handler  Handler          // 处理函数
	isParam  bool             // 是否是参数节点
}

// Router 基数树路由器
type Router struct {
	root map[string]*node // 不同方法的根节点
}

// NewRouter 创建新的路由器
func NewRouter() *Router {
	return &Router{
		root: make(map[string]*node),
	}
}

// addRoute 添加路由
func (r *Router) addRoute(method, path string, handler Handler) {
	if r.root[method] == nil {
		r.root[method] = &node{
			path:     "",
			children: make(map[string]*node),
		}
	}

	root := r.root[method]
	parts := parsePath(path)

	for _, part := range parts {
		child := root.children[part]
		if child == nil {
			child = &node{
				path:     part,
				children: make(map[string]*node),
				isParam:  part[0] == ':',
			}
			root.children[part] = child
		}
		root = child
	}
	root.handler = handler
}

// getRoute 获取路由和处理函数
func (r *Router) getRoute(method, path string) (*node, map[string]string) {
	params := make(map[string]string)
	root := r.root[method]
	if root == nil {
		return nil, nil
	}

	parts := parsePath(path)
	for _, part := range parts {
		found := false
		for _, child := range root.children {
			if child.isParam {
				params[child.path[1:]] = part
				root = child
				found = true
				break
			} else if child.path == part {
				root = child
				found = true
				break
			}
		}
		if !found {
			return nil, nil
		}
	}

	return root, params
}

// parsePath 解析路径为部分
func parsePath(path string) []string {
	parts := strings.Split(path, "/")
	result := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// Engine 引擎：串联 Router、中间件与对象池
type Engine struct {
	router   *Router
	handlers []Handler  // 全局中间件
	pool     sync.Pool  // 对象池，用于复用 Context
}

// New 创建新的引擎
func New() *Engine {
	engine := &Engine{
		router: NewRouter(),
	}

	// 初始化对象池
	engine.pool.New = func() interface{} {
		return &Context{}
	}

	return engine
}

// Use 添加全局中间件
func (e *Engine) Use(handlers ...Handler) {
	e.handlers = append(e.handlers, handlers...)
}

// GET 注册 GET 请求
func (e *Engine) GET(path string, handler Handler) {
	e.router.addRoute("GET", path, handler)
}

// POST 注册 POST 请求
func (e *Engine) POST(path string, handler Handler) {
	e.router.addRoute("POST", path, handler)
}

// ServeHTTP 实现 http.Handler 接口（使用对象池复用 Context）
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	node, params := e.router.getRoute(r.Method, r.URL.Path)
	if node == nil || node.handler == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 Not Found")
		return
	}

	// 从对象池获取 Context
	c := e.pool.Get().(*Context)
	c.reset() // 重置状态
	c.Writer = w
	c.Request = r
	c.Params = params
	c.index = -1
	c.handlers = append(e.handlers, node.handler)

	// 执行中间件链
	c.Next()

	// 归还到对象池
	e.pool.Put(c)
}

// Run 启动服务器
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}
