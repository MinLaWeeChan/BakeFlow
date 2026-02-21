import { createContext, useState, useEffect } from 'react'

export const LanguageContext = createContext({
  lang: 'en',
  setLang: () => {},
})

const STORAGE_KEY = 'bf_ui_lang'

export function LanguageProvider({ children }) {
  // Always start with 'en' to match SSR output and avoid hydration mismatch
  const [lang, setLang] = useState('en')

  // After hydration, sync from localStorage (client-only)
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored === 'en' || stored === 'my') {
        setLang(stored)
      }
    } catch (e) {
      // ignore
    }
  }, [])

  // Persist language changes to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, lang)
    } catch (e) {
      // ignore
    }
  }, [lang])

  return (
    <LanguageContext.Provider value={{ lang, setLang }}>
      {children}
    </LanguageContext.Provider>
  )
}
