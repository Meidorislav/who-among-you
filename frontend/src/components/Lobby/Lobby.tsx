import { useParams } from 'react-router'
import { useSession } from '../../contexts/SessionContext'

export const Lobby = () => {
  const { code } = useParams<{ code: string }>()
  const { session } = useSession()

  return (
    <main style={{ padding: 32, textAlign: 'center' }}>
      <h1>Lobby {code}</h1>
      <p>
        {session && session.code === code
          ? `Welcome, ${session.player.nickname}`
          : 'No session for this lobby (join flow coming soon).'}
      </p>
    </main>
  )
}
