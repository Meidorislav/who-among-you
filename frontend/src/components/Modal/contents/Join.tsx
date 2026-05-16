import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './Content.module.css'

type JoinProps = {
  name: string
  onNameChange: (value: string) => void
}

export const Join = ({ name, onNameChange }: JoinProps) => {
  const { t } = useTranslation()
  const [code, setCode] = useState('')
  const canJoin = name.trim().length > 0 && code.trim().length > 0

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

      <button className={styles.submit} disabled={!canJoin}>
        {t('join.submit')}
      </button>
    </>
  )
}
