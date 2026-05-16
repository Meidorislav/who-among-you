import styles from './Content.module.css'

const THEMES = [
  { id: 'sunset', name: 'Sunset' },
  { id: 'dark', name: 'Midnight' },
]

export const Theme = () => {
  return (
    <>
      <p className={styles.text}>Pick a vibe.</p>
      <div className={styles.list}>
        {THEMES.map((t) => (
          <button key={t.id} className={styles.option}>
            <span>{t.name}</span>
            {t.id === 'sunset' && <span className={styles.badge}>Active</span>}
          </button>
        ))}
      </div>
    </>
  )
}
