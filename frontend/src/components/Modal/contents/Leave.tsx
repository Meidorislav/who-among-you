import { useTranslation } from 'react-i18next'
import styles from './Content.module.css'
import leaveStyles from './Leave.module.css'

type LeaveProps = {
  onConfirm: () => void
  onCancel: () => void
}

export const Leave = ({ onConfirm, onCancel }: LeaveProps) => {
  const { t } = useTranslation()

  return (
    <>
      <p className={styles.text}>{t('leave.body')}</p>
      <div className={leaveStyles.actions}>
        <button className={leaveStyles.confirm} onClick={onConfirm}>
          {t('leave.confirm')}
        </button>
        <button className={leaveStyles.cancel} onClick={onCancel}>
          {t('leave.cancel')}
        </button>
      </div>
    </>
  )
}
