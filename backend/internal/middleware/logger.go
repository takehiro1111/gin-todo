package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var sensitiveFields = []string{
	"password",
	"token",
	"secret",
	"api_key",
	"apikey",
	"authorization",
}

// maskSensitiveData はJSONボディ内の機密情報をマスキング
func maskSensitiveData(body string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// JSONパース失敗時はそのまま返す
		return body
	}

	// 再帰的にマスキング
	masked := maskRecursive(data)

	// JSON文字列に戻す
	result, err := json.Marshal(masked)
	if err != nil {
		return body
	}
	return string(result)
}

// maskRecursive は再帰的に機密フィールドをマスキング
func maskRecursive(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// オブジェクト（マップ）の場合
		for key, value := range v {
			// キーが機密フィールドかチェック
			if isSensitiveField(key) {
				v[key] = "***MASKED***"
				continue
			}
			// 機密フィールドでなければ再帰的に処理
			v[key] = maskRecursive(value)
		}
		return v

	case []interface{}:
		// 配列の場合
		for i, item := range v {
			v[i] = maskRecursive(item)
		}
		return v

	default:
		// プリミティブ型（string, number, boolなど）はそのまま
		return v
	}
}

// isSensitiveField はフィールド名が機密情報かチェック
func isSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if lowerField == strings.ToLower(sensitive) {
			return true
		}
	}
	return false
}

// shouldLogBody はContent-Typeをチェックしてログ出力すべきか判定
func shouldLogBody(contentType string) bool {
	allowedTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"text/plain",
	}

	for _, allowedType := range allowedTypes {
		if strings.Contains(strings.ToLower(contentType), allowedType) {
			return true
		}
	}
	return false
}

// TimeFunc は時刻を取得する関数の型（例: time.Now）
type TimeFunc func() time.Time

// Option はLoggerConfigを設定するための関数型
// この型の関数は、LoggerConfigを受け取って設定を変更する
type Option func(*LoggerConfig)

// 時刻取得関数を設定するOptionを返す
// 使い方: NewLoggerConfig("info", "json", WithTimeFunc(time.Now))
func WithTimeFunc(f TimeFunc) Option {
	return func(l *LoggerConfig) {
		l.TimeNow = f
	}
}

// ログ出力先を設定するOptionを返す
// 使い方: NewLoggerConfig("info", "json", WithWriter(os.Stdout))
func WithWriter(w io.Writer) Option {
	return func(l *LoggerConfig) {
		l.Writer = w
	}
}

// WithSkipPaths は特定のパスをログから除外するOptionを返す
// ヘルスチェックエンドポイントなど、ログ不要なパスを指定
func WithSkipPaths(paths []string) Option {
	return func(l *LoggerConfig) {
		l.SkipPaths = paths
	}
}

// WithEnableRequestBody はリクエストボディのログ出力を有効化
// 注意: 機密情報が含まれる可能性があるため慎重に使用
func WithEnableRequestBody(enable bool) Option {
	return func(l *LoggerConfig) {
		l.EnableRequestBody = enable
	}
}

func WithUUIDGenerator(gen UUIDGenerator) Option {
	return func(l *LoggerConfig) {
		l.UUIDGen = gen
	}
}

type LoggerConfig struct {
	Level             string
	Format            string    // "json" or "text"
	TimeNow           TimeFunc  // 時刻取得関数（テスト時にモック可能）
	Writer            io.Writer // ログ出力先（アクセスログのため標準出力としてos.Stdoutを設定）
	SkipPaths         []string  // ログをスキップするパス
	EnableRequestBody bool      // リクエストボディをログ出力するか
	UUIDGen           UUIDGenerator
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewLogger(loggerConfig *LoggerConfig) *slog.Logger {
	level := parseLevel(loggerConfig.Level)
	opts := &slog.HandlerOptions{Level: level}

	writer := loggerConfig.Writer

	var handler slog.Handler
	switch strings.ToLower(loggerConfig.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler)
}

// NewLoggerConfig はLoggerConfigを生成する（Functional Optionsパターン）
//
// 使い方の例:
//
//	// デフォルト値を使用
//	config := NewLoggerConfig("info", "json")
//
//	// 出力先だけカスタマイズ
//	config := NewLoggerConfig("info", "json", WithWriter(os.Stdout))
//
//	// 複数のオプションを指定
//	config := NewLoggerConfig("info", "json",
//	    WithTimeFunc(mockTimeFunc),
//	    WithWriter(customWriter),
//	)
func NewLoggerConfig(level, format string, opts ...Option) *LoggerConfig {
	// ステップ1: デフォルト値でconfigを初期化
	config := &LoggerConfig{
		Level:             level,
		Format:            format,
		TimeNow:           time.Now,
		Writer:            os.Stdout,
		EnableRequestBody: false,
		UUIDGen:           &RealUUIDGenerator{},
		SkipPaths:         []string{},
	}

	// ステップ2: 渡されたオプション関数を順番に実行
	// opts = [
	//   func(l *LoggerConfig) { l.TimeNow = time.Now },
	//   func(l *LoggerConfig) { l.Writer = os.Stdout },
	// ]
	for _, opt := range opts {
		// optは「func(*LoggerConfig)」型の関数
		// opt(config) を実行すると、関数内部の代入処理が実行される
		//
		// 例: opt = func(l *LoggerConfig) { l.Writer = os.Stdout }
		//     opt(config) → config.Writer = os.Stdout が実行される
		opt(config)
	}

	return config
}

// 指定されたパスをスキップすべきか判定
func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// HTTPステータスコードに応じたログレベルを返す
func getLogLevel(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError
	case statusCode >= 400:
		// 404は頻発するためErrorで通知させたくない。
		return slog.LevelWarn
	case statusCode >= 300:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

type UUIDGenerator interface {
	Generate() string
}

type RealUUIDGenerator struct{}

func (g *RealUUIDGenerator) Generate() string {
	return uuid.New().String()
}

func LoggerMiddleware(loggerConfig *LoggerConfig) gin.HandlerFunc {
	slogger := NewLogger(loggerConfig)

	return func(c *gin.Context) {
		// スキップ対象パスに設定済みのエンドポイントはログを出力しない
		if shouldSkip(c.Request.URL.Path, loggerConfig.SkipPaths) {
			c.Next()
			return
		}

		start := loggerConfig.TimeNow()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 分散トレーシングのためのリクエストIDの取得 or デフォルト値設定
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// クライアントが渡していない場合は自動生成
			requestID = loggerConfig.UUIDGen.Generate()
			c.Header("X-Request-ID", requestID)
		}

		var requestBody string
		if loggerConfig.EnableRequestBody {
			contentType := c.GetHeader("Content-Type")
			if shouldLogBody(contentType) {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil && len(bodyBytes) > 0 {
					// ボディ全体を復元(後続のhandlerのため。)
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					// マスキング処理
					requestBody = maskSensitiveData(string(bodyBytes))
				}
			}
		}

		reqFields := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", requestID),
		}

		if loggerConfig.EnableRequestBody && requestBody != "" {
			reqFields = append(reqFields, slog.String("request_body", requestBody))
		}

		slogger.Info("[Req]", reqFields...)

		// アプリケーション側の処理を実行
		c.Next()

		// 処理直後の時間を取得したいためインスタンス化
		latency := loggerConfig.TimeNow().Sub(start)

		// ステータスコードに応じたログレベルの設定で出力
		statusCode := c.Writer.Status()
		logLevel := getLogLevel(statusCode)

		header := c.Writer.Header()
		resFields := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", statusCode),
			slog.String("latency", fmt.Sprintf("%.2fms", latency.Seconds()*1000)),
			slog.String("ip", c.ClientIP()),
			slog.Int("response_size", c.Writer.Size()),
			slog.String("request_id", requestID),
			slog.Any("response_header", header),
		}

		var errorMsgs []string
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				errorMsgs = append(errorMsgs, e.Error())
			}
			resFields = append(resFields, slog.Any("errors", errorMsgs))
		}

		// レスポンスのステータスコードに応じてレベルを動的に変更するため。
		slogger.Log(c.Request.Context(), logLevel, "[Res]", resFields...)
	}
}
