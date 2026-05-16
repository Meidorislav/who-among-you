import type {
  CreateLobbyResponse,
  JoinLobbyResponse,
} from './types'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const postJSON = async <T>(path: string, body: unknown): Promise<T> => {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    let message = text || res.statusText
    try {
      const parsed = JSON.parse(text)
      if (parsed?.error) message = parsed.error
    } catch {
      // body wasn't JSON, keep raw text
    }
    throw new ApiError(res.status, message)
  }
  return res.json() as Promise<T>
}

export const createLobby = (nickname: string) =>
  postJSON<CreateLobbyResponse>('/api/lobby', { nickname })

export const joinLobby = (nickname: string, code: string) =>
  postJSON<JoinLobbyResponse>('/api/lobby/join', { nickname, code })

export const socketURL = (code: string, playerId: string) => {
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const params = new URLSearchParams({ code, player_id: playerId })
  return `${scheme}//${window.location.host}/ws?${params}`
}
