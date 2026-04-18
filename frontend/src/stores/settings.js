import { defineStore } from 'pinia'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    language: localStorage.getItem('language') || 'en',
    direction: localStorage.getItem('direction') || 'ltr'
  }),
  actions: {
    setLanguage(lang) {
      this.language = lang
      this.direction = lang === 'ar' ? 'rtl' : 'ltr'
      localStorage.setItem('language', this.language)
      localStorage.setItem('direction', this.direction)
      
      // Update document direction immediately
      document.documentElement.dir = this.direction
      document.documentElement.lang = this.language
    },
    initSettings() {
      document.documentElement.dir = this.direction
      document.documentElement.lang = this.language
    }
  }
})
