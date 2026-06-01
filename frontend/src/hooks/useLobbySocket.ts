import { useCallback, useEffect, useRef, useState } from 'react'
import { socketURL } from '../api/client'
import type { ClientMessage, LobbySnapshot, ServerEvent } from '../api/types'

export type ConnectionState = 'connecting' | 'open' | 'closed'

export type RoundData = {
  round: number
  total: number
  questionEn: string
  questionRu: string
  deadlineMs: number | null
  roundDurationMs: number | null
  playerIds: string[]
  phase: 'voting' | 'results'
  votes: Record<string, number>
  scores: Record<string, number>
  winners: string[]
  nextReady: string[]
}

export type LobbySocket = {
  lobby: LobbySnapshot | null
  connection: ConnectionState
  countdownDeadline: number | null
  gameStarted: boolean
  gameRound: RoundData | null
  myVote: string | null
  finalScores: Record<string, number> | null
  setReady: (ready: boolean) => void
  updateSettings: (questionCount: number, roundDurationSeconds: number) => void
  kickPlayer: (targetPlayerId: string) => void
  vote: (targetPlayerId: string) => void
  nextRound: () => void
}

export const useLobbySocket = (
  code: string,
  playerId: string,
  initialLobby?: LobbySnapshot,
): LobbySocket => {
  const [lobby, setLobby] = useState<LobbySnapshot | null>(initialLobby ?? null)
  const [connection, setConnection] = useState<ConnectionState>('connecting')
  const [countdownDeadline, setCountdownDeadline] = useState<number | null>(null)
  const [gameStarted, setGameStarted] = useState(false)
  const [gameRound, setGameRound] = useState<RoundData | null>(null)
  const [myVote, setMyVote] = useState<string | null>(null)
  const [finalScores, setFinalScores] = useState<Record<string, number> | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (!code || !playerId) return

    const url = socketURL(code, playerId)
    const ws = new WebSocket(url)
    wsRef.current = ws
    setConnection('connecting')
    setCountdownDeadline(null)
    setGameStarted(false)
    console.debug('[ws] opening', url)

    ws.onopen = () => {
      console.debug('[ws] open', code, playerId)
      setConnection('open')
    }
    ws.onclose = (e) => {
      console.debug('[ws] close', code, playerId, e.code, e.reason)
      setConnection('closed')
    }
    ws.onerror = (e) => {
      console.debug('[ws] error', code, playerId, e)
      setConnection('closed')
    }

    ws.onmessage = (e) => {
      console.debug('[ws] message', code, playerId, e.data)
      let event: ServerEvent
      try {
        event = JSON.parse(e.data)
      } catch {
        return
      }
      switch (event.type) {
        case 'lobby_state':
          setLobby(event.lobby)
          break
        case 'countdown_started':
          setCountdownDeadline(event.deadline)
          break
        case 'countdown_cancelled':
          setCountdownDeadline(null)
          break
        case 'game_started':
          setCountdownDeadline(null)
          setGameStarted(true)
          break
        case 'round_started':
          setGameRound({
            round: event.round,
            total: event.total,
            questionEn: event.question_en,
            questionRu: event.question_ru,
            deadlineMs: event.deadline > 0 ? event.deadline * 1000 : null,
            roundDurationMs:
              event.round_duration_seconds > 0
                ? event.round_duration_seconds * 1000
                : null,
            playerIds: event.players,
            phase: 'voting',
            votes: {},
            scores: {},
            winners: [],
            nextReady: [],
          })
          setMyVote(null)
          break
        case 'round_ended':
          setGameRound((prev) =>
            prev
              ? {
                  ...prev,
                  phase: 'results',
                  votes: event.votes,
                  scores: event.scores,
                  winners: event.winners,
                  nextReady: event.next_ready,
                }
              : null,
          )
          break
        case 'next_round_state':
          setGameRound((prev) =>
            prev && prev.round === event.round
              ? { ...prev, nextReady: event.next_ready }
              : prev,
          )
          break
        case 'game_finished':
          setFinalScores(event.scores)
          setGameRound(null)
          break
      }
    }

    return () => {
      console.debug('[ws] cleanup', code, playerId)
      ws.onopen = null
      ws.onclose = null
      ws.onerror = null
      ws.onmessage = null
      ws.close()
      if (wsRef.current === ws) wsRef.current = null
    }
  }, [code, playerId])

  const send = useCallback((msg: ClientMessage) => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg))
    }
  }, [])

  const setReady = useCallback(
    (ready: boolean) => send({ type: 'set_ready', ready }),
    [send],
  )
  const updateSettings = useCallback(
    (questionCount: number, roundDurationSeconds: number) =>
      send({
        type: 'update_settings',
        question_count: questionCount,
        round_duration_seconds: roundDurationSeconds,
      }),
    [send],
  )
  const kickPlayer = useCallback(
    (targetPlayerId: string) => send({ type: 'kick_player', target_player_id: targetPlayerId }),
    [send],
  )
  const vote = useCallback(
    (targetPlayerId: string) => {
      send({ type: 'vote', target_player_id: targetPlayerId })
      setMyVote(targetPlayerId)
    },
    [send],
  )
  const nextRound = useCallback(() => send({ type: 'next_round' }), [send])

  return {
    lobby,
    connection,
    countdownDeadline,
    gameStarted,
    gameRound,
    myVote,
    finalScores,
    setReady,
    updateSettings,
    kickPlayer,
    vote,
    nextRound,
  }
}
