package middleware

import (
	"github.com/gin-gonic/gin"
)

func SetSecurityResponseHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// レスポンスヘッダーをセット
		// 2年間HTTPSを強制。サブドメインにも適用
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		// iframeへの埋め込みを全て禁止（クリックジャッキング対策
		c.Header("X-Frame-Options", "DENY")
		// ブラウザがContent-Typeを勝手に推測しない（XSS対策）
		c.Header("X-Content-Type-Options", "nosniff")
		// 自ドメインのリソースのみ読み込み許可（外部スクリプト等を遮断）
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}
