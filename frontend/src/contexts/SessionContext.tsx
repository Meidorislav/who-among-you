import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import type { Player } from '../api/types'

export type Session = {
  player: Player
  code: string
}

type SessionContextValue = {
  session: Session | null
  setSession: (session: Session | null) => void
}

const STORAGE_KEY = 'session'
const SessionContext = createContext<SessionContextValue | null>(null)

const readInitialSession = (): Session | null => {
  if (typeof window === 'undefined') return null
  const raw = window.localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Session
    if (parsed?.player?.player_id && parsed?.code) return parsed
  } catch {
    // fall through
  }
  return null
}

export const SessionProvider = ({ children }: { children: ReactNode }) => {
  const [session, setSessionState] = useState<Session | null>(readInitialSession)

  useEffect(() => {
    if (session) {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
    } else {
      window.localStorage.removeItem(STORAGE_KEY)
    }
  }, [session])

  const setSession = useCallback((next: Session | null) => setSessionState(next), [])

  return (
    <SessionContext.Provider value={{ session, setSession }}>
      {children}
    </SessionContext.Provider>
  )
}

export const useSession = () => {
  const ctx = useContext(SessionContext)
  if (!ctx) throw new Error('useSession must be used inside <SessionProvider>')
  return ctx
}
