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

	// writeWait はクライアントへの書き込み操作のタイムアウト。
	// Ping送信やメッセージ送信時にSetWriteDeadlineで使用する。
	writeWait = 10 * time.Second
)

// Client はWebSocket接続を持つ個別のクライアントを表す。
// チャットルーム機能やユーザー情報の紐付け時に拡張予定。
type Client struct {
	conn *websocket.Conn
}

// sendPeriodicPing はpingPeriod間隔でクライアントにPingメッセージを送信する。
// クライアントはPingを受信すると自動的にPongを返し、PongHandlerでReadDeadlineが延長される。
// 一定時間Pongが返ってこない場合はreadTimeout超過により接続が切断される。
// doneチャネルがクローズされると、このgoroutineも終了する。
func sendPeriodicPing(conn *websocket.Conn, timeProvider utils.TimeProvider, done chan struct{}) {
	// pingPeriod間隔で発火するTickerを生成
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		conn.Close()
	}()

	for {
		select {
		// 読み取り側が終了した場合、Ping送信も停止する
		case <-done:
			return

		// 定期的にPingを送信して接続の死活を確認する
		case <-pingTicker.C:
			conn.SetWriteDeadline(timeProvider.Now().Add(writeWait))
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("write:", err)
				return
			}
			log.Println("ping sent")
		}
	}
}
