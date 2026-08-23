package middleware

import (
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

func RequestID(newID func() string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = newID()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func CORS(origins map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && origins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID,Idempotency-Key")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func Recover(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if value := recover(); value != nil {
				logger.Error("panic recovered", zap.Any("panic", value), zap.ByteString("stack", debug.Stack()), zap.String("request_id", CurrentRequestID(c)))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "服务暂时不可用", "field_errors": []any{}, "request_id": CurrentRequestID(c)})
			}
		}()
		c.Next()
	}
}

func CurrentRequestID(c *gin.Context) string {
	value, _ := c.Get(RequestIDKey)
	id, _ := value.(string)
	return id
}

type bucket struct {
	Tokens  float64
	Updated time.Time
}
type Limiter struct {
	mu          sync.Mutex
	buckets     map[string]bucket
	Rate, Burst float64
}

func NewLimiter(rate, burst float64) *Limiter {
	return &Limiter{buckets: map[string]bucket{}, Rate: rate, Burst: burst}
}
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		l.mu.Lock()
		b := l.buckets[key]
		if b.Updated.IsZero() {
			b = bucket{Tokens: l.Burst, Updated: now}
		}
		elapsed := now.Sub(b.Updated).Seconds()
		b.Tokens = min(l.Burst, b.Tokens+elapsed*l.Rate)
		allowed := b.Tokens >= 1
		if allowed {
			b.Tokens--
		}
		b.Updated = now
		l.buckets[key] = b
		l.mu.Unlock()
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁", "field_errors": []any{}, "request_id": CurrentRequestID(c)})
			return
		}
		c.Next()
	}
}
