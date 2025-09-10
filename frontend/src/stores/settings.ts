import { defineStore } from "pinia";

// Settings interface
interface Settings {
  endpoint: string;
  apiKey: string;
  knowledgeBaseId: string;
}

// Default settings
const defaultSettings: Settings = {
  endpoint: import.meta.env.VITE_IS_DOCKER ? "" : "http://localhost:8080",
  apiKey: "",
  knowledgeBaseId: "",
};

export const useSettingsStore = defineStore("settings", {
  state: () => ({
    // Load settings from localStorage; prefer new key then fallback to old key; default if missing
    settings: (() => {
      const NEW_KEY = "WeKnoRust_settings";
      const OLD_KEY = "WeKnora_settings";
      const raw = localStorage.getItem(NEW_KEY) ?? localStorage.getItem(OLD_KEY);
      try {
        return raw ? JSON.parse(raw) : { ...defaultSettings };
      } catch {
        return { ...defaultSettings };
      }
    })(),
  }),

  actions: {
    // Save settings
    saveSettings(settings: Settings) {
      this.settings = { ...settings };
      // Save to localStorage (write both keys for backward compatibility)
      localStorage.setItem("WeKnoRust_settings", JSON.stringify(this.settings));
      localStorage.setItem("WeKnora_settings", JSON.stringify(this.settings));
    },

    // Get settings
    getSettings(): Settings {
      return this.settings;
    },

    // Get API endpoint
    getEndpoint(): string {
      return this.settings.endpoint || defaultSettings.endpoint;
    },

    // Get API Key
    getApiKey(): string {
      return this.settings.apiKey;
    },

    // Get knowledge base ID
    getKnowledgeBaseId(): string {
      return this.settings.knowledgeBaseId;
    },
  },
}); 