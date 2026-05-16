import styles from './Header.module.css'
import logoSunset from '../../assets/logo_sunset.png'
import logoMidnight from '../../assets/logo_midnight.png'
import type { ModalType } from '../../App'
import { useTheme } from '../../contexts/ThemeContext'

const LOGOS = {
  sunset: logoSunset,
  midnight: logoMidnight,
}

type HeaderProps = {
  onOpen: (modal: ModalType) => void
}

export const Header = ({ onOpen }: HeaderProps) => {
  const { theme } = useTheme()

  return (
    <header className={styles.header}>
      <img src={LOGOS[theme]} alt="Who Among You?" className={styles.logo} />
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
