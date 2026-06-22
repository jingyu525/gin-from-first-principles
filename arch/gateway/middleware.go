package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
)

// ==================== 1. 日志中间件 ====================

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 执行后续处理
		c.Next()

		// 记录日志
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		log.Printf("[Gateway] %s %s %d %v", method, path, statusCode, duration)
	}
}

// ==================== 2. 限流中间件 ====================

// RateLimiter 限流器（令牌桶算法）
type RateLimiter struct {
	rate     float64   // 每秒允许的请求数
	capacity int       // 桶容量
	tokens   float64   // 当前令牌数
	lastTime time.Time // 上次更新时间
	mu       sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate float64, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		lastTime: time.Now(),
	}
}

// Allow 判断是否允许请求
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastTime).Seconds()

	// 补充令牌
	r.tokens += elapsed * r.rate
	if r.tokens > float64(r.capacity) {
		r.tokens = float64(r.capacity)
	}
	r.lastTime = now

	// 判断是否有足够令牌
	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rate float64, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// ==================== 3. 熔断中间件 ====================

// CircuitBreakerMiddleware 熔断中间件
func CircuitBreakerMiddleware() gin.HandlerFunc {
	// 创建熔断器
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "GatewayCircuitBreaker",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 连续失败 5 次，触发熔断
			return counts.ConsecutiveFailures >= 5
		},
	})

	return func(c *gin.Context) {
		// 执行请求（通过熔断器保护）
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next()
			// 如果状态码 >= 500，认为是失败
			if c.Writer.Status() >= 500 {
				return nil, http.ErrAbortHandler
			}
			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Circuit breaker is open",
			})
		}
	}
}

// ==================== 4. 链路追踪中间件 ====================

// TracingMiddleware 链路追踪中间件（简化版）
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或提取 TraceID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 生成 SpanID
		spanID := generateSpanID()

		// 注入到请求头
		c.Request.Header.Set("X-Trace-ID", traceID)
		c.Request.Header.Set("X-Span-ID", spanID)

		// 设置到 Gin 上下文
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)

		// 记录日志
		log.Printf("[Tracing] TraceID=%s SpanID=%s Method=%s Path=%s",
			traceID, spanID, c.Request.Method, c.Request.URL.Path)

		c.Next()
	}
}

// generateTraceID 生成 TraceID
func generateTraceID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// generateSpanID 生成 SpanID
func generateSpanID() string {
	return randomString(16)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ==================== 5. Prometheus 指标中间件 ====================

var (
	// 请求计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// 请求延迟直方图
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// init 初始化 Prometheus 指标
func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// PrometheusMiddleware Prometheus 指标中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 执行后续处理
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsEndpoint 暴露 Prometheus 指标端点
func MetricsEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}
