import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Route, Routes, useNavigate } from 'react-router'
import { Background } from './components/Background/Background'
import { Header } from './components/Header/Header'
import { Home } from './components/Home/Home'
import { Footer } from './components/Footer/Footer'
import { Lobby } from './components/Lobby/Lobby'
import { Modal } from './components/Modal/Modal'
import { AboutGame } from './components/Modal/contents/AboutGame'
import { Language } from './components/Modal/contents/Language'
import { Theme } from './components/Modal/contents/Theme'
import { Join } from './components/Modal/contents/Join'
import { Leave } from './components/Modal/contents/Leave'
import { useSession } from './contexts/SessionContext'
import { ApiError, createLobby } from './api/client'

export type ModalType = 'about' | 'language' | 'theme' | 'join' | 'leave'

function App() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { setSession } = useSession()

  const [name, setName] = useState('')
  const [modal, setModal] = useState<ModalType | null>(null)
  const [error, setError] = useState<string | null>(null)

  const closeModal = () => setModal(null)

  const handleLeave = () => {
    setSession(null)
    navigate('/')
    closeModal()
  }

  const handleCreate = async () => {
    const nickname = name.trim()
    if (!nickname) {
      setError(t('errors.nicknameRequired'))
      return
    }
    setError(null)
    try {
      const { code, player, lobby } = await createLobby(nickname)
      setSession({ player, code })
      navigate(`/lobby/${code}`, {
        state: {
          initialLobby: lobby,
        },
      })
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : t('errors.network')
      setError(msg)
    }
  }

  return (
    <>
      <Background />
      <Routes>
        <Route
          path="/"
          element={
            <>
              <Header onOpen={setModal} />
              <Home
                name={name}
                onNameChange={setName}
                onCreate={handleCreate}
                onJoin={() => setModal('join')}
                error={error}
              />
              <Footer />
            </>
          }
        />
        <Route path="/lobby/:code" element={
          <>
            <Header onOpen={setModal} onLogoClick={() => setModal('leave')} />
            <Lobby />
            <Footer />
          </>
        }
        />
      </Routes>

      {modal && (
        <Modal title={t(`modal.titles.${modal}`)} onClose={closeModal}>
          {modal === 'about' && <AboutGame />}
          {modal === 'language' && <Language />}
          {modal === 'theme' && <Theme />}
          {modal === 'join' && (
            <Join name={name} onNameChange={setName} onSuccess={closeModal} />
          )}
          {modal === 'leave' && (
            <Leave onConfirm={handleLeave} onCancel={closeModal} />
          )}
        </Modal>
      )}
    </>
  )
}

export default App
