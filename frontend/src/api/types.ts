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
  | { type: 'round_started'; [key: string]: unknown }
  | { type: 'round_ended'; [key: string]: unknown }
  | { type: 'game_finished'; [key: string]: unknown }

export type ClientMessage =
  | { type: 'set_ready'; ready: boolean }
  | { type: 'vote'; target_player_id: string }
