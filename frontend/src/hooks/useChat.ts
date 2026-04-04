import { tokenStorage } from '@/utils/tokenStorage'
import { useEffect, useRef, useState } from 'react'

interface ChatMessage {
  user: string
  message: string
}

/**
 * WebSocket チャットに接続するカスタムフック。
 * マウント時に接続し、アンマウント時に切断する。
 */
export function useChat() {
  const wsRef = useRef<WebSocket | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    const ws = new WebSocket(
      `${import.meta.env.VITE_API_BASE_URL?.replace('http', 'ws')}/api/ws/chat`
    )
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
      // 接続直後、最初のメッセージとして JWT を送信して認証
      const token = tokenStorage.getAccess()
      if (token) ws.send(token)
    }

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data as string) as ChatMessage
      setMessages((prev) => [...prev, msg])
    }

    ws.onclose = () => {
      setIsConnected(false)
      wsRef.current = null
    }

    // アンマウント時に切断する
    return () => ws.close()
  }, [])

  const sendMessage = (text: string) => {
    wsRef.current?.send(text)
  }

  return { messages, sendMessage, isConnected }
}
