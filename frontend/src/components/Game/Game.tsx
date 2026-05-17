import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Player } from '../../api/types'
import type { RoundData } from '../../hooks/useLobbySocket'
import styles from './Game.module.css'

const ROUND_TOTAL_MS = 45_000

type GameScreenProps = {
  players: Player[]
  selfId: string
  gameRound: RoundData | null
  myVote: string | null
  finalScores: Record<string, number> | null
  vote: (targetPlayerId: string) => void
  onLeave: () => void
}

export const GameScreen = ({
  players,
  selfId,
  gameRound,
  myVote,
  finalScores,
  vote,
  onLeave,
}: GameScreenProps) => {
  const nickname = (id: string) =>
    players.find((p) => p.player_id === id)?.nickname ?? id.slice(0, 8)

  if (finalScores) {
    return <FinishedPhase players={players} scores={finalScores} onLeave={onLeave} />
  }

  if (!gameRound) {
    return (
      <main className={styles.game}>
        <span className={styles.loading}>···</span>
      </main>
    )
  }

  if (gameRound.phase === 'results') {
    return <ResultsPhase round={gameRound} players={players} nickname={nickname} />
  }

  return (
    <VotingPhase
      round={gameRound}
      selfId={selfId}
      myVote={myVote}
      vote={vote}
      nickname={nickname}
    />
  )
}

// ---------------------------------------------------------------------------
// Voting
// ---------------------------------------------------------------------------

type VotingPhaseProps = {
  round: RoundData
  selfId: string
  myVote: string | null
  vote: (id: string) => void
  nickname: (id: string) => string
}

const VotingPhase = ({ round, selfId, myVote, vote, nickname }: VotingPhaseProps) => {
  const { t } = useTranslation()
  const hasVoted = myVote !== null

  return (
    <main className={styles.game}>
      <div className={styles.badge}>
        {t('game.round', { round: round.round, total: round.total })}
      </div>

      <div className={styles.questionBox}>
        <p className={styles.question}>{round.question}</p>
      </div>

      <TimerBar key={round.deadlineMs} deadlineMs={round.deadlineMs} totalMs={ROUND_TOTAL_MS} />

      {hasVoted && <p className={styles.hint}>{t('game.waitingForVotes')}</p>}

      <ul className={styles.playerList}>
        {round.playerIds.map((id) => {
          const isSelf = id === selfId
          const isChosen = myVote === id
          return (
            <li key={id}>
              <button
                className={styles.voteBtn}
                data-chosen={isChosen}
                disabled={hasVoted}
                onClick={() => vote(id)}
              >
                <span className={styles.voteName}>{nickname(id)}</span>
                {isSelf && <span className={styles.tag}>{t('lobby.you')}</span>}
                {isChosen && <span className={styles.check}>✓</span>}
              </button>
            </li>
          )
        })}
      </ul>
    </main>
  )
}

// ---------------------------------------------------------------------------
// Results
// ---------------------------------------------------------------------------

type ResultsPhaseProps = {
  round: RoundData
  players: Player[]
  nickname: (id: string) => string
}

const ResultsPhase = ({ round, players, nickname }: ResultsPhaseProps) => {
  const { t } = useTranslation()

  const sorted = [...players].sort(
    (a, b) => (round.votes[b.player_id] ?? 0) - (round.votes[a.player_id] ?? 0),
  )

  return (
    <main className={styles.game}>
      <div className={styles.badge}>
        {t('game.roundOver', { round: round.round })}
      </div>

      <div className={styles.questionBox}>
        <p className={styles.question}>{round.question}</p>
      </div>

      {round.winners.length > 0 ? (
        <p className={styles.winners}>
          {t('game.winnersThis')} <strong>{round.winners.map(nickname).join(', ')}</strong>
        </p>
      ) : (
        <p className={styles.hint}>{t('game.noVotes')}</p>
      )}

      <ul className={styles.resultList}>
        {sorted.map((p) => {
          const votes = round.votes[p.player_id] ?? 0
          const score = round.scores[p.player_id] ?? 0
          const isWinner = round.winners.includes(p.player_id)
          return (
            <li key={p.player_id} className={styles.resultRow} data-winner={isWinner}>
              <span className={styles.resultName}>{p.nickname}</span>
              <span className={styles.resultVotes}>{votes > 0 ? `+${votes}` : '—'}</span>
              <span className={styles.resultScore}>{score}</span>
            </li>
          )
        })}
      </ul>

      <p className={styles.hint}>{t('game.nextRound')}</p>
    </main>
  )
}

// ---------------------------------------------------------------------------
// Game finished
// ---------------------------------------------------------------------------

type FinishedPhaseProps = {
  players: Player[]
  scores: Record<string, number>
  onLeave: () => void
}

const FinishedPhase = ({ players, scores, onLeave }: FinishedPhaseProps) => {
  const { t } = useTranslation()

  const ranked = [...players].sort(
    (a, b) => (scores[b.player_id] ?? 0) - (scores[a.player_id] ?? 0),
  )

  return (
    <main className={styles.game}>
      <div className={styles.badge}>{t('game.gameOver')}</div>

      <ul className={styles.resultList}>
        {ranked.map((p, i) => (
          <li key={p.player_id} className={styles.resultRow} data-winner={i === 0}>
            <span className={styles.rank}>#{i + 1}</span>
            <span className={styles.resultName}>{p.nickname}</span>
            <span className={styles.resultScore}>{scores[p.player_id] ?? 0}</span>
          </li>
        ))}
      </ul>

      <button className={styles.leaveBtn} onClick={onLeave}>
        {t('game.goHome')}
      </button>
    </main>
  )
}

// ---------------------------------------------------------------------------
// Timer bar
// ---------------------------------------------------------------------------

const TimerBar = ({ deadlineMs, totalMs }: { deadlineMs: number; totalMs: number }) => {
  const [initial] = useState(() => {
    const remaining = Math.max(0, deadlineMs - Date.now())
    return { remaining, pct: remaining / totalMs }
  })

  const [secondsLeft, setSecondsLeft] = useState(() => Math.ceil(initial.remaining / 1000))

  useEffect(() => {
    const id = setInterval(() => {
      setSecondsLeft(Math.max(0, Math.ceil((deadlineMs - Date.now()) / 1000)))
    }, 200)
    return () => clearInterval(id)
  }, [deadlineMs])

  const [draining, setDraining] = useState(false)
  useEffect(() => {
    const raf = requestAnimationFrame(() => setDraining(true))
    return () => cancelAnimationFrame(raf)
  }, [])

  const scaleX = draining ? 0 : initial.pct
  const transitionMs = draining ? initial.remaining : 0

  return (
    <div className={styles.timerWrap}>
      <div className={styles.timerTrack}>
        <div
          className={styles.timerFill}
          style={{
            transform: `scaleX(${scaleX})`,
            transitionDuration: `${transitionMs}ms`,
          }}
        />
      </div>
      <span className={styles.timerSeconds}>{secondsLeft}s</span>
    </div>
  )
}
