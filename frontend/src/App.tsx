import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Background } from './components/Background/Background'
import { Header } from './components/Header/Header'
import { Home } from './components/Home/Home'
import { Footer } from './components/Footer/Footer'
import { Modal } from './components/Modal/Modal'
import { AboutGame } from './components/Modal/contents/AboutGame'
import { Language } from './components/Modal/contents/Language'
import { Theme } from './components/Modal/contents/Theme'
import { Join } from './components/Modal/contents/Join'

export type ModalType = 'about' | 'language' | 'theme' | 'join'

function App() {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [modal, setModal] = useState<ModalType | null>(null)

  const closeModal = () => setModal(null)

  return (
    <>
      <Background />
      <Header onOpen={setModal} />
      <Home name={name} onNameChange={setName} onJoin={() => setModal('join')} />
      <Footer />

      {modal && (
        <Modal title={t(`modal.titles.${modal}`)} onClose={closeModal}>
          {modal === 'about' && <AboutGame />}
          {modal === 'language' && <Language />}
          {modal === 'theme' && <Theme />}
          {modal === 'join' && <Join name={name} onNameChange={setName} />}
        </Modal>
      )}
    </>
  )
}

export default App
