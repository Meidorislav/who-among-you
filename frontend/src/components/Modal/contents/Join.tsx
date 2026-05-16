import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ApiError, joinLobby } from '../../../api/client'
import { useSession } from '../../../contexts/SessionContext'
import { useNavigate } from 'react-router'
import styles from './Content.module.css'

type JoinProps = {
  name: string
  onNameChange: (value: string) => void
  onSuccess: () => void
}

export const Join = ({ name, onNameChange, onSuccess }: JoinProps) => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { setSession } = useSession()

  const [code, setCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const canJoin = name.trim().length > 0 && code.trim().length > 0 && !submitting

  const handleJoin = async () => {
    if (!canJoin) return
    setError(null)
    setSubmitting(true)
    try {
      const { player, lobby } = await joinLobby(name.trim(), code.trim())
      setSession({ player, code: lobby.code })
      onSuccess()
      navigate(`/lobby/${lobby.code}`)
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : t('errors.network')
      setError(msg)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>{t('join.nicknameLabel')}</span>
        <input
          className={styles.input}
          type="text"
          placeholder={t('join.nicknamePlaceholder')}
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
        />
      </div>

      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>{t('join.codeLabel')}</span>
        <input
          className={styles.input}
          type="text"
          placeholder={t('join.codePlaceholder')}
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
        />
      </div>

      {error && <p className={styles.error}>{error}</p>}

      <button
        className={styles.submit}
        disabled={!canJoin}
        onClick={handleJoin}
      >
        {t('join.submit')}
      </button>
    </>
  )
}
