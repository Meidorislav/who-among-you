import styles from './Header.module.css'

export const Header = () => {
  return (
    <header className={styles.header}>
      <h1 className={styles.logo}>Who Among You?</h1>
      <div className={styles.navigation_container}>
        <button className={styles.navigation_button}>About Game</button>
        <button className={styles.navigation_button}>Language</button>
        <button className={styles.navigation_button}>Theme</button>
        <button className={styles.navigation_button}>GitHub</button>
      </div>
    </header>
  )
}
