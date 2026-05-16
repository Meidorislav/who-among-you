import { useState } from 'react'
import styles from './Content.module.css'

type JoinProps = {
  name: string
  onNameChange: (value: string) => void
}

export const Join = ({ name, onNameChange }: JoinProps) => {
  const [code, setCode] = useState('')
  const canJoin = name.trim().length > 0 && code.trim().length > 0

  return (
    <>
      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>Nickname</span>
        <input
          className={styles.input}
          type="text"
          placeholder="Type your name..."
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
        />
      </div>

      <div className={styles.inputWrap}>
        <span className={styles.inputLabel}>Room code</span>
        <input
          className={styles.input}
          type="text"
          placeholder="ABCDEF"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
        />
      </div>

      <button className={styles.submit} disabled={!canJoin}>
        Join room
      </button>
    </>
  )
}
