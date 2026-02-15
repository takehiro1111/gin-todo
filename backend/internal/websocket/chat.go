package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

// Message はWebSocketで送受信されるメッセージを表す。
// Type はwebsocket.TextMessageやwebsocket.BinaryMessage等のメッセージ種別。
type Message struct {
	Type    int
	Message []byte
}

// Hub はWebSocket接続の中央管理を行う構造体。
// 全クライアントの接続状態を保持し、メッセージのブロードキャスト配信を担当する。
type Hub struct {
	// clients は現在接続中のWebSocketコネクションを管理するマップ
	clients map[*websocket.Conn]bool
	// broadcast は全クライアントへ配信するメッセージを受け取るチャネル
	broadcast chan Message
	// register は新しいクライアント接続の登録要求を受け取るチャネル
	register chan *websocket.Conn
	// unregister はクライアント切断時の登録解除要求を受け取るチャネル
	unregister chan *websocket.Conn
	// mu はclientsマップへの並行アクセスを保護するRWMutex
	mu sync.RWMutex
	// timeProvider はテスト時にモック可能な時刻取得インターフェース
	timeProvider utils.TimeProvider
	// jwtProvider はWebSocket接続時のJWT認証に使用するプロバイダー
	jwtProvider utils.JWTProvider
}

// ChatMessage はブロードキャスト時に全クライアントへ送信されるJSONメッセージを表す。
type ChatMessage struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// NewHub はHubのインスタンスを生成する。
// broadcastチャネルは256のバッファを持ち、一時的なメッセージの集中に対応する。
func NewHub(timeProvider utils.TimeProvider, jwtProvider utils.JWTProvider) *Hub {
	return &Hub{
		clients:      make(map[*websocket.Conn]bool),
		broadcast:    make(chan Message, 256),
		register:     make(chan *websocket.Conn),
		unregister:   make(chan *websocket.Conn),
		timeProvider: timeProvider,
		jwtProvider:  jwtProvider,
	}
}

// Run はHubのメインイベントループ。
// main.goで `go hub.Run()` としてgoroutineで起動し、アプリケーション終了まで常駐する。
// register/unregister/broadcastの3つのチャネルを監視し、
// クライアントの接続管理とメッセージ配信を行う。
func (h *Hub) Run() {
	for {
		select {
		// 新しいクライアントが接続した際にclientsマップへ登録する
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()

		// クライアントが切断した際にclientsマップから削除し、コネクションを閉じる
		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()

		// broadcastチャネルからメッセージを受け取り、全クライアントへ配信する。
		// 書き込みに失敗したクライアントは切断済みと判断し、マップから削除する。
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(message.Type, message.Message)
				if err != nil {
					// client側がコネクションを切ったときなど、プロセスやサーバーを止めたくないためlogで対応
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// ChatServer godoc
// @Summary      WebSocketチャット接続
// @Description  JWT認証付きのWebSocketチャットエンドポイント。クエリパラメータでJWTトークンを渡してHTTP接続をWebSocketにアップグレードする。接続後はリアルタイムでメッセージの送受信が可能。
// @Tags         websocket
// @Param        token query string true "JWTアクセストークン"
// @Success      101 {string} string "WebSocket接続確立"
// @Failure      401 {object} utils.ErrorResponse "認証エラー"
// @Router       /api/ws/chat [get]
func (h *Hub) ChatServer(c *gin.Context) {
	jwtToken := c.Query("token")
	claims, err := h.jwtProvider.VerifyJWT(jwtToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// HTTP → WebSocket へのプロトコルアップグレード設定
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// productionのプロダクトではパフォーマンス気をつける。
			allowedOrigins := []string{
				"http://localhost:3000",
			}
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	// doneチャネルはこの接続のライフサイクルを管理し、Ping送信goroutineの停止に使用する
	done := make(chan struct{})

	// HTTPリクエストをWebSocketコネクションにアップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %+v\n", err)
		return
	}
	defer func() {
		conn.Close()
		close(done)          // sendPeriodicPingのgoroutineを停止させる
		h.unregister <- conn // Hubのclientsマップから削除する
	}()

	// Hubにこのクライアントを登録し、ブロードキャスト配信の対象にする
	h.register <- conn

	// 初期接続時のReadDeadlineを設定（readTimeout以内にメッセージが来なければタイムアウト）
	conn.SetReadDeadline(h.timeProvider.Now().Add(readTimeout))

	// Pong受信時にReadDeadlineを延長し、接続を維持する
	conn.SetPongHandler(func(string) error {
		log.Println("pong received")
		conn.SetReadDeadline(h.timeProvider.Now().Add(readTimeout))
		return nil
	})

	// 定期的にPingを送信するgoroutineを起動（接続の死活監視）
	go sendPeriodicPing(conn, h.timeProvider, done)

	// クライアントからのメッセージを読み取り、Hubのbroadcastチャネルへ送信する
	// この関数は接続が切断されるまでブロックする
	h.readAndBroadcast(conn, claims, done)
}

// readAndBroadcast はクライアントからのメッセージを読み取り、Hubのbroadcastチャネルに送信する。
// 接続が切断されるかエラーが発生するまでループし続ける。
// ChatServerから呼び出され、各クライアント接続ごとに1つ実行される。
func (h *Hub) readAndBroadcast(conn *websocket.Conn, claims *utils.JWTClaims, done chan struct{}) {
	for {
		// クライアントからメッセージを読み取る（メッセージが届くまでブロック）
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			// 正常な切断（ブラウザ閉じる等）と異常切断を区別してログ出力
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Client disconnected normally: %v", err)
			} else {
				log.Printf("ReadMessage Error: %+v\n", err)
			}
			break
		}

		chatMsg, err := json.Marshal(ChatMessage{
			User:    claims.UserName,
			Message: string(msg),
		})

		if err != nil {
			log.Printf("failed encoding json: %+v\n", err)
			continue
		}

		// Hub.Run()のbroadcast処理に委譲し、全クライアントへ配信する
		h.broadcast <- Message{Type: mt, Message: chatMsg}
	}
}
