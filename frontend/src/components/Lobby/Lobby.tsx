import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Navigate, useLocation, useParams } from 'react-router'
import type { LobbySnapshot, Player } from '../../api/types'
import { useSession, type Session } from '../../contexts/SessionContext'
import { useLobbySocket } from '../../hooks/useLobbySocket'
import styles from './Lobby.module.css'

const COUNTDOWN_TOTAL_MS = 5000
// viewBox is 100x100, centered at (50, 50). r=46 with stroke-width=8 puts the
// stroke's outer edge exactly at r=50 — flush against the SVG boundary, which
// (since the SVG fills the container's content box with inset:0) lands exactly
// at the inner edge of the container's frame border.
const RING_RADIUS = 46
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS

export const Lobby = () => {
  const { code } = useParams<{ code: string }>()
  const { session } = useSession()
  const location = useLocation()
  const initialLobby = (location.state as { initialLobby?: LobbySnapshot } | null)
    ?.initialLobby

  if (!code || !session || session.code !== code) {
    return <Navigate to="/" replace />
  }

  return <LobbyView code={code} session={session} initialLobby={initialLobby} />
}

type LobbyViewProps = {
  code: string
  session: Session
  initialLobby?: LobbySnapshot
}

const LobbyView = ({ code, session, initialLobby }: LobbyViewProps) => {
  const { t } = useTranslation()
  const { lobby, connection, countdownDeadline, gameStarted, setReady } =
    useLobbySocket(code, session.player.player_id, initialLobby)

  const players: Player[] = lobby?.players ?? [session.player]
  const me = players.find((p) => p.player_id === session.player.player_id)
  const isReady = me?.ready ?? false
  const canToggle = connection === 'open' && !gameStarted
  const inCountdown = countdownDeadline !== null

  return (
    <main className={styles.lobby}>
      <div className={styles.codeWrap}>
        <span className={styles.codeLabel}>{t('lobby.codeLabel')}</span>
        <h1 className={styles.code}>{code}</h1>
      </div>

      <p className={styles.hint}>{t('lobby.hint')}</p>

      {connection !== 'open' && (
        <p className={styles.status} data-tone="warn">
          {t(`lobby.connection.${connection}`)}
        </p>
      )}

      {inCountdown && (
        <CountdownRing key={countdownDeadline} deadline={countdownDeadline} />
      )}

      <ul className={styles.players}>
        {players.map((p) => {
          const isSelf = p.player_id === session.player.player_id
          return (
            <li
              key={p.player_id}
              className={styles.player}
              data-ready={p.ready}
              data-self={isSelf}
            >
              <span className={styles.nickname}>
                {p.nickname}
                {isSelf && <span className={styles.youTag}>{t('lobby.you')}</span>}
              </span>
              <span
                className={styles.dot}
                aria-label={p.ready ? t('lobby.ready') : t('lobby.waiting')}
              />
            </li>
          )
        })}
      </ul>

      <button
        className={styles.readyBtn}
        data-ready={isReady}
        data-countdown={inCountdown}
        disabled={!canToggle}
        onClick={() => setReady(!isReady)}
      >
        {inCountdown
          ? t('lobby.cancelStart')
          : isReady
            ? t('lobby.unready')
            : t('lobby.ready')}
      </button>

      {gameStarted && (
        <p className={styles.status} data-tone="ok">
          {t('lobby.gameStarted')}
        </p>
      )}
    </main>
  )
}

type CountdownRingProps = { deadline: number }

// CountdownRing animates the visual progress via a CSS transition on
// stroke-dashoffset — browser-native, GPU-friendly, no per-frame JS work.
// The number ticker uses a cheap interval since we only need second-level
// resolution. Parent re-mounts this on every new deadline (via key=deadline)
// so the captured initial snapshot stays consistent for the whole animation.
const CountdownRing = ({ deadline }: CountdownRingProps) => {
  const { t } = useTranslation()

  const [initial] = useState(() => {
    const remaining = Math.max(0, deadline - Date.now())
    const offset = (1 - remaining / COUNTDOWN_TOTAL_MS) * RING_CIRCUMFERENCE
    return { remaining, offset }
  })

  const [secondsLeft, setSecondsLeft] = useState(() =>
    Math.ceil(initial.remaining / 1000),
  )

  useEffect(() => {
    const update = () => {
      setSecondsLeft(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)))
    }
    const id = window.setInterval(update, 200)
    return () => window.clearInterval(id)
  }, [deadline])

  // Two-frame paint: first render holds the explicit "from" offset, next frame
  // flips to the empty-ring target so the CSS transition has something to
  // interpolate between.
  const [draining, setDraining] = useState(false)
  useEffect(() => {
    const raf = requestAnimationFrame(() => setDraining(true))
    return () => cancelAnimationFrame(raf)
  }, [])

  const dashOffset = draining ? RING_CIRCUMFERENCE : initial.offset
  const transitionMs = draining ? initial.remaining : 0

  return (
    <div className={styles.countdown}>
      <svg className={styles.ring} viewBox="0 0 100 100" aria-hidden="true">
        <circle
          className={styles.ringTrack}
          cx="50"
          cy="50"
          r={RING_RADIUS}
        />
        <circle
          className={styles.ringFill}
          cx="50"
          cy="50"
          r={RING_RADIUS}
          strokeDasharray={RING_CIRCUMFERENCE}
          strokeDashoffset={dashOffset}
          transform="rotate(-90 50 50)"
          style={{ transitionDuration: `${transitionMs}ms` }}
        />
      </svg>
      <div className={styles.countdownText}>
        <span className={styles.countdownLabel}>{t('lobby.startingIn')}</span>
        <span key={secondsLeft} className={styles.countdownNumber}>
          {secondsLeft}
        </span>
      </div>
    </div>
  )
}
