package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/takehiro1111/gin-todo/backend/internal/utils"
)

const (
	// クライアントからのPongまたはメッセージ受信を待つ最大時間。
	// この時間内にPongが返ってこない場合、接続が切れたと判断する。
	readTimeout = 60 * time.Second

	// pingPeriod はサーバーからクライアントへPingを送信する間隔。
	// readTimeoutの90%に設定し、タイムアウト前に必ずPingが届くようにするため9で割っている。
	pingPeriod = (readTimeout * 9) / 10

	// クライアントへの書き込み操作のタイムアウト。
	writeWait = 10 * time.Second

	// maxMessageSize はクライアントから受信するメッセージの最大サイズ（バイト）。
	maxMessageSize = 4096
)

// Client はWebSocket接続を持つ個別のクライアントを表す。
// sendチャネルを介して書き込みをwritePumpに集約し、並行書き込みを防止する。
type Client struct {
	conn *websocket.Conn
	send chan Message
}

// writePump はsendチャネルとPingTickerを単一のgoroutineで処理し、
// connへの書き込みを直列化する。各Clientにつき1つ起動される。
// sendチャネルがクローズされるか、書き込みエラーが発生すると終了する。
func (cl *Client) writePump(timeProvider utils.TimeProvider) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		cl.conn.Close()
	}()

	for {
		select {
		case message, ok := <-cl.send:
			if !ok {
				// Hubがsendチャネルを閉じた（unregisterまたはbroadcast時のバッファ満杯）
				// 基盤となるネットワーク接続の書き込み期限を設定します。
				cl.conn.SetWriteDeadline(timeProvider.Now().Add(writeWait))
				cl.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			cl.conn.SetWriteDeadline(timeProvider.Now().Add(writeWait))
			err := cl.conn.WriteMessage(message.Type, message.Message)
			if err != nil {
				log.Printf("writePump WriteMessage error: %v", err)
				return
			}

		case <-pingTicker.C:
			cl.conn.SetWriteDeadline(timeProvider.Now().Add(writeWait))
			err := cl.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Printf("writePump Ping error: %v", err)
				return
			}
			log.Println("ping sent")
		}
	}
}
