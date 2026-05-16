import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation()

  return (
    <header className={styles.header}>
      <img src={LOGOS[theme]} alt={t('header.logoAlt')} className={styles.logo} />
      <div className={styles.navigation_container}>
        <button className={styles.navigation_button} onClick={() => onOpen('about')}>
          {t('header.about')}
        </button>
        <button className={styles.navigation_button} onClick={() => onOpen('language')}>
          {t('header.language')}
        </button>
        <button className={styles.navigation_button} onClick={() => onOpen('theme')}>
          {t('header.theme')}
        </button>
        <a
          className={styles.navigation_button}
          href="https://github.com/Meidorislav/who-among-you"
          target="_blank"
          rel="noreferrer"
        >
          {t('header.github')}
        </a>
      </div>
    </header>
  )
}
