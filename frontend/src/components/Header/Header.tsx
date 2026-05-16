import styles from './Header.module.css'
import logo from '../../assets/logo.png'

export const Header = () => {
  return (
    <header className={styles.header}>
      <img src={logo} alt="Who Among You?" className={styles.logo} />
      <div className={styles.navigation_container}>
        <button className={styles.navigation_button}>About Game</button>
        <button className={styles.navigation_button}>Language</button>
        <button className={styles.navigation_button}>Theme</button>
        <button className={styles.navigation_button}>GitHub</button>
      </div>
    </header>
  )
}
