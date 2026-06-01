export type LobbyStatus = 'waiting' | 'playing' | 'finished'

export type Player = {
  nickname: string
  player_id: string
  ready: boolean
}

export type LobbySettings = {
  question_count: number
  round_duration_seconds: number
}

export type LobbySnapshot = {
  code: string
  status: LobbyStatus
  host_id: string
  settings: LobbySettings
  players: Player[]
}

export type CreateLobbyResponse = {
  code: string
  player: Player
  lobby: LobbySnapshot
}

export type JoinLobbyResponse = {
  player: Player
  lobby: LobbySnapshot
}

export type ServerEvent =
  | { type: 'lobby_state'; lobby: LobbySnapshot }
  | { type: 'countdown_started'; deadline: number }
  | { type: 'countdown_cancelled' }
  | { type: 'game_started' }
  | {
      type: 'round_started'
      round: number
      total: number
      question_en: string
      question_ru: string
      deadline: number // Unix seconds
      round_duration_seconds: number
      players: string[] // player UUIDs
    }
  | {
      type: 'round_ended'
      round: number
      votes: Record<string, number>
      scores: Record<string, number>
      winners: string[]
      next_ready: string[]
    }
  | { type: 'next_round_state'; round: number; next_ready: string[] }
  | { type: 'game_finished'; scores: Record<string, number> }

export type ClientMessage =
  | { type: 'set_ready'; ready: boolean }
  | { type: 'update_settings'; question_count: number; round_duration_seconds: number }
  | { type: 'kick_player'; target_player_id: string }
  | { type: 'vote'; target_player_id: string }
  | { type: 'next_round' }
