export type LobbyStatus = 'waiting' | 'playing' | 'finished'

export type Player = {
  nickname: string
  player_id: string
  ready: boolean
}

export type LobbySnapshot = {
  code: string
  status: LobbyStatus
  players: Player[]
}

export type CreateLobbyResponse = {
  code: string
  player: Player
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
      question: string
      deadline: number // Unix seconds
      players: string[] // player UUIDs
    }
  | {
      type: 'round_ended'
      round: number
      votes: Record<string, number>
      scores: Record<string, number>
      winners: string[]
    }
  | { type: 'game_finished'; scores: Record<string, number> }

export type ClientMessage =
  | { type: 'set_ready'; ready: boolean }
  | { type: 'vote'; target_player_id: string }
