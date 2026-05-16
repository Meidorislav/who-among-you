import { useTranslation } from 'react-i18next'
import styles from './Home.module.css'

type HomeProps = {
  name: string
  onNameChange: (value: string) => void
  onJoin: () => void
}

export const Home = ({ name, onNameChange, onJoin }: HomeProps) => {
  const { t } = useTranslation()
  const title = t('brand')

  return (
    <main className={styles.home}>
      <h1 className={styles.title}>
        {title.split('').map((char, i) => (
          <span
            key={i}
            className={styles.titleLetter}
            style={{ animationDelay: `${i * 0.05}s` }}
          >
            {char === ' ' ? ' ' : char}
          </span>
        ))}
      </h1>

      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>{t('home.nicknameLabel')}</span>
        <input
          className={styles.input}
          type="text"
          placeholder={t('home.nicknamePlaceholder')}
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
        />
      </div>

      <div className={styles.actions}>
        <button className={`${styles.button} ${styles.buttonPrimary}`}>
          {t('home.create')}
        </button>
        <button
          className={`${styles.button} ${styles.buttonSecondary}`}
          onClick={onJoin}
        >
          {t('home.join')}
        </button>
      </div>
    </main>
  )
}
