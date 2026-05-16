import { Trans, useTranslation } from 'react-i18next'
import styles from './Content.module.css'

export const AboutGame = () => {
  const { t } = useTranslation()
  return (
    <>
      <p className={styles.text}>
        <Trans i18nKey="about.p1" components={[<b />]} />
      </p>
      <p className={styles.text}>{t('about.p2')}</p>
      <p className={styles.text}>{t('about.p3')}</p>
      <p className={styles.text}>{t('about.p4')}</p>
    </>
  )
}
