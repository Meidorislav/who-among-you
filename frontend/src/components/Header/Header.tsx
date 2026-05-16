import styles from './Header.module.css'
import logo from '../../assets/logo.png'
import type { ModalType } from '../../App'

type HeaderProps = {
  onOpen: (modal: ModalType) => void
}

export const Header = ({ onOpen }: HeaderProps) => {
  return (
    <header className={styles.header}>
      <img src={logo} alt="Who Among You?" className={styles.logo} />
      <div className={styles.navigation_container}>
        <button className={styles.navigation_button} onClick={() => onOpen('about')}>
          About Game
        </button>
        <button className={styles.navigation_button} onClick={() => onOpen('language')}>
          Language
        </button>
        <button className={styles.navigation_button} onClick={() => onOpen('theme')}>
          Theme
        </button>
        <a
          className={styles.navigation_button}
          href="https://github.com/Meidorislav/who-among-you"
          target="_blank"
          rel="noreferrer"
        >
          GitHub
        </a>
      </div>
    </header>
  )
}
