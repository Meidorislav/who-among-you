import { useTranslation } from 'react-i18next'
import styles from './Content.module.css'
import { SUPPORTED_LANGUAGES, type Language as Lang } from '../../../i18n'

const LANGUAGE_NAMES: Record<Lang, string> = {
  en: 'English',
  ru: 'Русский',
}

export const Language = () => {
  const { i18n, t } = useTranslation()
  const current = (i18n.resolvedLanguage ?? i18n.language).split('-')[0] as Lang

  return (
    <>
      <p className={styles.text}>{t('language.prompt')}</p>
      <div className={styles.list}>
        {SUPPORTED_LANGUAGES.map((code) => (
          <button
            key={code}
            className={styles.option}
            onClick={() => i18n.changeLanguage(code)}
          >
            <span>{LANGUAGE_NAMES[code]}</span>
            {code === current && <span className={styles.badge}>{t('language.active')}</span>}
          </button>
        ))}
      </div>
    </>
  )
}
