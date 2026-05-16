import { useTranslation } from 'react-i18next'
import styles from './Content.module.css'
import { THEMES, useTheme } from '../../../contexts/ThemeContext'

export const Theme = () => {
  const { theme, setTheme } = useTheme()
  const { t } = useTranslation()

  return (
    <>
      <p className={styles.text}>{t('theme.prompt')}</p>
      <div className={styles.list}>
        {THEMES.map((th) => (
          <button
            key={th.id}
            className={styles.option}
            onClick={() => setTheme(th.id)}
          >
            <span>{t(`theme.${th.id}`)}</span>
            {th.id === theme && <span className={styles.badge}>{t('theme.active')}</span>}
          </button>
        ))}
      </div>
    </>
  )
}
