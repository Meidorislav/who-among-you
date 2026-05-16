import styles from './Content.module.css'

const LANGUAGES = [
  { code: 'en', name: 'English' },
  { code: 'ru', name: 'Русский' },
]

export const Language = () => {
  return (
    <>
      <p className={styles.text}>Choose your language.</p>
      <div className={styles.list}>
        {LANGUAGES.map((l) => (
          <button key={l.code} className={styles.option}>
            <span>{l.name}</span>
            {l.code === 'en' && <span className={styles.badge}>Active</span>}
          </button>
        ))}
      </div>
    </>
  )
}
