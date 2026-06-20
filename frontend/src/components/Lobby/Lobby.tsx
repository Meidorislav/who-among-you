import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Navigate, useLocation, useNavigate, useParams } from 'react-router'
import { fetchCategories } from '../../api/client'
import type { LobbySnapshot, Player } from '../../api/types'
import { useSession, type Session } from '../../contexts/SessionContext'
import { useLobbySocket } from '../../hooks/useLobbySocket'
import { GameScreen } from '../Game/Game'
import styles from './Lobby.module.css'

const COUNTDOWN_TOTAL_MS = 5000
const QUESTION_OPTIONS = [5, 10, 15, 20]
const ROUND_TIME_OPTIONS = [15, 30, 45, 60, 90]
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

  if (!code || !session) {
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
  const navigate = useNavigate()
  const { setSession } = useSession()
  const [availableCategories, setAvailableCategories] = useState<string[]>([])

  useEffect(() => {
    fetchCategories().then(setAvailableCategories)
  }, [])

  const {
    lobby,
    connection,
    countdownDeadline,
    gameStarted,
    setReady,
    updateSettings,
    kickPlayer,
    vote,
    nextRound,
    gameRound,
    myVote,
    finalScores,
  } = useLobbySocket(code, session.player.player_id, initialLobby)

  const handleLeave = useCallback(() => {
    setSession(null)
    navigate('/')
  }, [navigate, setSession])

  useEffect(() => {
    if (!lobby) return
    const stillInLobby = lobby.players.some((p) => p.player_id === session.player.player_id)
    if (!stillInLobby) {
      handleLeave()
    }
  }, [handleLeave, lobby, session.player.player_id])

  if (gameStarted || lobby?.status === 'playing') {
    return (
      <GameScreen
        players={lobby?.players ?? []}
        selfId={session.player.player_id}
        gameRound={gameRound}
        myVote={myVote}
        finalScores={finalScores}
        vote={vote}
        nextRound={nextRound}
        onLeave={handleLeave}
      />
    )
  }

  const players: Player[] = lobby?.players ?? [session.player]
  const me = players.find((p) => p.player_id === session.player.player_id)
  const isReady = me?.ready ?? false
  const isHost = lobby?.host_id === session.player.player_id
  const canToggle = connection === 'open' && !gameStarted
  const inCountdown = countdownDeadline !== null
  const settings = lobby?.settings ?? {
    question_count: 10,
    round_duration_seconds: 45,
    categories: [] as string[],
  }
  const selectedCategories = settings.categories ?? []
  const allSelected = selectedCategories.length === 0

  const isCategorySelected = (cat: string) => allSelected || selectedCategories.includes(cat)

  const toggleCategory = (cat: string) => {
    let next: string[]
    if (allSelected) {
      next = availableCategories.filter((c) => c !== cat)
    } else if (selectedCategories.includes(cat)) {
      if (selectedCategories.length === 1) return
      next = selectedCategories.filter((c) => c !== cat)
    } else {
      const added = [...selectedCategories, cat]
      next = added.length === availableCategories.length ? [] : added
    }
    updateSettings(settings.question_count, settings.round_duration_seconds, next)
  }

  const resetCategories = () => {
    updateSettings(settings.question_count, settings.round_duration_seconds, [])
  }

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

      {isHost && (
        <section className={styles.hostPanel} aria-label={t('lobby.hostPanel')}>
          <div className={styles.hostTitle}>{t('lobby.hostPanel')}</div>
          <label className={styles.setting}>
            <span>{t('lobby.questionCount')}</span>
            <select
              value={settings.question_count}
              disabled={connection !== 'open'}
              onChange={(event) =>
                updateSettings(Number(event.target.value), settings.round_duration_seconds, selectedCategories)
              }
            >
              {QUESTION_OPTIONS.map((count) => (
                <option key={count} value={count}>
                  {count}
                </option>
              ))}
              <option value={0}>{t('lobby.allQuestions')}</option>
            </select>
          </label>

          <label className={styles.setting}>
            <span>{t('lobby.roundTime')}</span>
            <select
              value={settings.round_duration_seconds}
              disabled={connection !== 'open'}
              onChange={(event) =>
                updateSettings(settings.question_count, Number(event.target.value), selectedCategories)
              }
            >
              {ROUND_TIME_OPTIONS.map((seconds) => (
                <option key={seconds} value={seconds}>
                  {t('lobby.seconds', { count: seconds })}
                </option>
              ))}
              <option value={0}>{t('lobby.noTimeLimit')}</option>
            </select>
          </label>

          {availableCategories.length > 0 && (
            <div className={styles.categoriesSection}>
              <div className={styles.categoriesHeader}>
                <span className={styles.categoriesLabel}>{t('lobby.categories')}</span>
                {!allSelected && (
                  <button
                    type="button"
                    className={styles.resetCategoriesBtn}
                    disabled={connection !== 'open'}
                    onClick={resetCategories}
                  >
                    {t('lobby.selectAll')}
                  </button>
                )}
              </div>
              <div className={styles.categoryGrid}>
                {availableCategories.map((cat) => (
                  <label key={cat} className={styles.categoryCheckbox}>
                    <input
                      type="checkbox"
                      checked={isCategorySelected(cat)}
                      disabled={connection !== 'open'}
                      onChange={() => toggleCategory(cat)}
                    />
                    {t(`lobby.categoryNames.${cat}`, { defaultValue: cat })}
                  </label>
                ))}
              </div>
            </div>
          )}
        </section>
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
                {p.player_id === lobby?.host_id && (
                  <span className={styles.hostTag}>{t('lobby.host')}</span>
                )}
              </span>
              <span className={styles.playerActions}>
                {isHost && !isSelf && (
                  <button
                    className={styles.kickBtn}
                    type="button"
                    disabled={connection !== 'open'}
                    onClick={() => kickPlayer(p.player_id)}
                  >
                    {t('lobby.kick')}
                  </button>
                )}
                <span
                  className={styles.dot}
                  aria-label={p.ready ? t('lobby.ready') : t('lobby.waiting')}
                />
              </span>
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
