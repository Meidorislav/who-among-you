import styles from './Home.module.css'

const TITLE = 'Who Among You?'

export const Home = () => {
  return (
    <main className={styles.home}>
      <h1 className={styles.title}>
        {TITLE.split('').map((char, i) => (
          <span
            key={i}
            className={styles.titleLetter}
            style={{ animationDelay: `${i * 0.05}s` }}
          >
            {char === ' ' ? ' ' : char}
          </span>
        ))}
      </h1>

      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>Nickname</span>
        <input
          className={styles.input}
          type="text"
          placeholder="Type your name..."
        />
      </div>

      <div className={styles.actions}>
        <button className={`${styles.button} ${styles.buttonPrimary}`}>
          Create
        </button>
        <button className={`${styles.button} ${styles.buttonSecondary}`}>
          Join
        </button>
      </div>
    </main>
  )
}
