import { useTranslation } from 'react-i18next'
import styles from './Footer.module.css'

export const Footer = () => {
  const { t } = useTranslation()
  return (
    <footer className={styles.footer}>
      <p className={styles.credits}>{t('footer.credits')}</p>
    </footer>
  )
}
