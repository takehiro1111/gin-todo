package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
	"golang.org/x/time/rate"
)

const (
	ErrMsgBadRequest     = "Bad Request"
	ErrMsgTooManyRequest = "Too Many Requests"
)

// テスト時に現在時刻に依存しないモックを作成するためにinterfaceを定義
type timeProvider interface {
	Now() time.Time
}

// 本番環境用の時刻プロバイダー実装
type RateLimitRealTimeProvider struct{}

func (*RateLimitRealTimeProvider) Now() time.Time {
	return time.Now()
}

// レート制限とクリーンアップの設定
type RateLimitConfig struct {
	TimeProvider      timeProvider
	CleanupInterval   time.Duration
	InactiveThreshold time.Duration
	RateLimit         time.Duration
	Burst             int
}

// クリーンアップ判定のため最終アクセス時刻を記録
type limiterInfo struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// 複数のリクエスト間での状態を共有するためにグローバル変数を使用
var (
	mu          sync.Mutex
	limiters    = make(map[string]*limiterInfo)
	timeConfig  *RateLimitConfig
	stopCleanup = make(chan struct{}) // メモリ0バイトのため、struct{}を使用している
	cleanupOnce sync.Once
)

// テスト用のDI
func InitRateLimiter(cfg *RateLimitConfig) {
	timeConfig = cfg
}

// メモリリークを防ぐため定期的にクリーンアップを実行
func StartCleanup() {
	// StartCleanup()が複数回呼ばれても、goroutineは1つだけ起動
	cleanupOnce.Do(func() {
		if timeConfig == nil {
			timeConfig = &RateLimitConfig{
				TimeProvider:      &RateLimitRealTimeProvider{},
				CleanupInterval:   10 * time.Minute,
				InactiveThreshold: -30 * time.Minute,
				RateLimit:         time.Second,
				Burst:             100, // 100回の上限
			}
		}

		// メモリリークを防ぐため定期的にクリーンアップを実行
		ticker := time.NewTicker(timeConfig.CleanupInterval)
		go func() {
			// stopCleanup を受信するとticker.Stop()が実行される
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					cleanupLimiters()
				case <-stopCleanup:
					return
				}
			}
		}()
	})
}

func cleanupLimiters() {
	mu.Lock()
	defer mu.Unlock()

	// メモリの中で30分以上アクセスのないIPのratelimiter情報は削除される。
	// 30分という値は便宜上キリの良い閾値として設定しています。
	threshold := timeConfig.TimeProvider.Now().Add(timeConfig.InactiveThreshold) // 30min
	count := 0

	for ip, info := range limiters {
		// 30分より前と境界も削除する。
		if info.lastAccess.Before(threshold) || info.lastAccess.Equal(threshold) {
			delete(limiters, ip)
			count++
		}
	}

	if count > 0 {
		log.Printf("Cleaned up %d inactive rate limiters\n", count)
	}
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	// 同一IPで既存のratelimitが設定されている場合は取得のみ
	info, exists := limiters[ip]
	if exists {
		limiters[ip].lastAccess = timeConfig.TimeProvider.Now()
		return info.limiter
	}

	limit := rate.Every(timeConfig.RateLimit)
	burst := timeConfig.Burst

	// 6秒毎に1トークン蓄積 / 10トークン分のburst可能
	limiter := rate.NewLimiter(limit, burst)
	limiters[ip] = &limiterInfo{
		limiter:    limiter,
		lastAccess: timeConfig.TimeProvider.Now(),
	}

	return limiter
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// プロキシやロードバランサーを経由する場合を想定して
		ip := c.Request.Header.Get("X-Forwarded-For")
		if ip == "" {
			var err error
			ip, _, err = net.SplitHostPort(c.Request.RemoteAddr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, utils.ErrorResponse{
					Success:   false,
					Message:   "bad request",
					Error:     ErrMsgBadRequest,
					Timestamp: time.Now(),
				})
				return
			}
		}

		limiter := getLimiter(ip)

		if !limiter.Allow() {
			// クライアントに再試行可能な時間を通知するためヘッダーを設定。
			c.Header("Retry-After", "60")
			log.Printf("Rate limit exceeded for IP: %s\n", ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, utils.ErrorResponse{
				Success:   false,
				Message:   "too many requests",
				Error:     ErrMsgTooManyRequest,
				Timestamp: time.Now(),
			})
			return
		}

		c.Next()
	}
}
