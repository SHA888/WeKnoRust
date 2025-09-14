// src/utils/request
import axios from "axios";
import { generateRandomString } from "./index";

// Read settings from localStorage (prefer new key, fallback to old)
function getSettings() {
  const settingsStr = localStorage.getItem("WeKnoRust_settings") ?? localStorage.getItem("WeKnora_settings");
  if (settingsStr) {
    try {
      return JSON.parse(settingsStr);
    } catch (e) {
      console.error("Failed to parse settings:", e);
    }
  }
  return {
    endpoint: import.meta.env.VITE_IS_DOCKER ? "" : "http://localhost:8080",
    apiKey: "",
    knowledgeBaseId: "",
  };
}

// API base URL — prefer endpoint from settings
const settings = getSettings();
const BASE_URL = settings.endpoint;

// Test data (optional)
let testData: {
  tenant: {
    id: number;
    name: string;
    api_key: string;
  };
  knowledge_bases: Array<{
    id: string;
    name: string;
    description: string;
  }>;
} | null = null;

// Create Axios instance
const instance = axios.create({
  baseURL: BASE_URL, // Use configured API base URL
  timeout: 30000, // Request timeout
  headers: {
    "Content-Type": "application/json",
    "X-Request-ID": `${generateRandomString(12)}`,
  },
});

// Set test data
export function setTestData(data: typeof testData) {
  testData = data;
  if (data) {
    // Prefer ApiKey from settings; fallback to test data
    const apiKey = settings.apiKey || (data?.tenant?.api_key || "");
    if (apiKey) {
      instance.defaults.headers["X-API-Key"] = apiKey;
    }
  }
}

// Get test data
export function getTestData() {
  return testData;
}

instance.interceptors.request.use(
  (config) => {
    // Before each request: check for updated settings
    const currentSettings = getSettings();
    
    // Update baseURL if changed
    if (currentSettings.endpoint && config.baseURL !== currentSettings.endpoint) {
      config.baseURL = currentSettings.endpoint;
    }
    
    // Update API Key if provided
    if (currentSettings.apiKey) {
      (config.headers ||= {})["X-API-Key"] = currentSettings.apiKey;
    }
    
    (config.headers ||= {})["X-Request-ID"] = `${generateRandomString(12)}`;
    return config;
  },
  (error) => {}
);

instance.interceptors.response.use(
  (response) => {
    // Handle by HTTP status code
    const { status, data } = response;
    if (status === 200 || status === 201) {
      return data;
    } else {
      return Promise.reject(data);
    }
  },
  (error: any) => {
    if (!error.response) {
      return Promise.reject({ message: "Network error, please check your connection" });
    }
    const { data } = error.response;
    return Promise.reject(data);
  }
);

export function get(url: string) {
  return instance.get(url);
}

export async function getDown(url: string) {
  let res = await instance.get(url, {
    responseType: "blob",
  });
  return res
}

export function postUpload(url: string, data = {}) {
  return instance.post(url, data, {
    headers: {
      "Content-Type": "multipart/form-data",
      "X-Request-ID": `${generateRandomString(12)}`,
    },
  });
}

export function postChat(url: string, data = {}) {
  return instance.post(url, data, {
    headers: {
      "Content-Type": "text/event-stream;charset=utf-8",
      "X-Request-ID": `${generateRandomString(12)}`,
    },
  });
}

export function post(url: string, data = {}) {
  return instance.post(url, data);
}

export function put(url: string, data = {}) {
  return instance.put(url, data);
}

export function del(url: string) {
  return instance.delete(url);
}
