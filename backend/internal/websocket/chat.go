package websocket

import (
	"encoding/json"
	"log"
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
	// clients は現在接続中のClientを管理するマップ
	clients map[*Client]bool
	// broadcast は全クライアントへ配信するメッセージを受け取るチャネル
	broadcast chan Message
	// register は新しいクライアント接続の登録要求を受け取るチャネル
	register chan *Client
	// unregister はクライアント切断時の登録解除要求を受け取るチャネル
	unregister chan *Client
	// mu はclientsマップへの並行アクセスを保護するRWMutex
	mu sync.RWMutex
	// timeProvider はテスト時にモック可能な時刻取得インターフェース
	timeProvider utils.TimeProvider
	// jwtProvider はWebSocket接続時のJWT認証に使用するプロバイダー
	jwtProvider utils.JWTProvider

	upgrader websocket.Upgrader
}

// ChatMessage はブロードキャスト時に全クライアントへ送信されるJSONメッセージを表す。
type ChatMessage struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// NewHub はHubのインスタンスを生成する。
// broadcastチャネルは256のバッファを持ち、一時的なメッセージの集中に対応する。
func NewHub(timeProvider utils.TimeProvider, jwtProvider utils.JWTProvider, upgrader websocket.Upgrader) *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan Message, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		timeProvider: timeProvider,
		jwtProvider:  jwtProvider,
		upgrader:     upgrader,
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
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		// クライアントが切断した際にclientsマップから削除し、sendチャネルを閉じる。
		// sendチャネルのクローズによりwritePumpが終了し、connもクローズされる。
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		// broadcastチャネルからメッセージを受け取り、全クライアントのsendチャネルへ配信する。
		// sendバッファが満杯のクライアントは遅延と判断し、切断する。
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// sendバッファが満杯 → 遅いクライアントを切断
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// ChatServer godoc
// @Summary      WebSocketチャット接続
// @Description  JWT認証付きのWebSocketチャットエンドポイント。接続後、最初のメッセージでJWTトークンを送信して認証する。認証成功後はリアルタイムでメッセージの送受信が可能。
// @Tags         websocket
// @Success      101 {string} string "WebSocket接続確立"
// @Failure      401 {object} utils.ErrorResponse "認証エラー"
// @Router       /api/ws/chat [get]
func (h *Hub) ChatServer(c *gin.Context) {
	// HTTPリクエストをWebSocketコネクションにアップグレード
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %+v\n", err)
		return
	}

	// Upgrade後、最初のメッセージでトークンを受け取る
	_, tokenMsg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	claims, err := h.jwtProvider.VerifyJWT(string(tokenMsg))
	if err != nil {
		// Upgrade後はHTTPレスポンスを返せないため、WebSocketメッセージでエラーを通知
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"unauthorized"}`))
		conn.Close()
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan Message, 256),
	}
	h.register <- client

	defer func() {
		h.unregister <- client
	}()

	// メッセージの最大サイズの設定
	conn.SetReadLimit(maxMessageSize)

	// 初期接続時のReadDeadlineを設定（readTimeout以内にメッセージが来なければタイムアウト）
	conn.SetReadDeadline(h.timeProvider.Now().Add(readTimeout))

	// Pong受信時にReadDeadlineを延長し、接続を維持する
	conn.SetPongHandler(func(string) error {
		log.Println("pong received")
		conn.SetReadDeadline(h.timeProvider.Now().Add(readTimeout))
		return nil
	})

	// 書き込みを直列化するgoroutineを起動（broadcastメッセージ送信 + Ping送信）
	go client.writePump(h.timeProvider)

	// クライアントからのメッセージを読み取り、Hubのbroadcastチャネルへ送信する
	// この関数は接続が切断されるまでブロックする
	h.readAndBroadcast(conn, claims)
}

// readAndBroadcast はクライアントからのメッセージを読み取り、Hubのbroadcastチャネルに送信する。
// 接続が切断されるかエラーが発生するまでループし続ける。
// ChatServerから呼び出され、各クライアント接続ごとに1つ実行される。
func (h *Hub) readAndBroadcast(conn *websocket.Conn, claims *utils.JWTClaims) {
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
