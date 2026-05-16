import styles from './Content.module.css'
import { THEMES, useTheme } from '../../../contexts/ThemeContext'

export const Theme = () => {
  const { theme, setTheme } = useTheme()

  return (
    <>
      <p className={styles.text}>Pick a vibe.</p>
      <div className={styles.list}>
        {THEMES.map((t) => (
          <button
            key={t.id}
            className={styles.option}
            onClick={() => setTheme(t.id)}
          >
            <span>{t.name}</span>
            {t.id === theme && <span className={styles.badge}>Active</span>}
          </button>
        ))}
      </div>
    </>
  )
}
