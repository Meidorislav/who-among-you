import styles from './Content.module.css'

export const AboutGame = () => {
  return (
    <>
      <p className={styles.text}>
        <b>Who Among You?</b> is a cozy little party game to get to know each
        other better — or just have a great evening with old friends.
      </p>
      <p className={styles.text}>
        Every round a question pops up, and you pick who fits the description
        best. Voting for yourself? Totally allowed. After each round there's
        time to chat about it and share the stories behind your answers.
      </p>
      <p className={styles.text}>
        Whoever gets the most votes that round earns a point — but honestly,
        there are no winners or losers here. The whole point is to learn
        something new about each other.
      </p>
      <p className={styles.text}>
        No app, no install — just open the link and play.
      </p>
    </>
  )
}
