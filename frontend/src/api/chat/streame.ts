import { fetchEventSource } from '@microsoft/fetch-event-source'
import { ref, type Ref, onUnmounted, nextTick } from 'vue'
import { generateRandomString } from '@/utils/index';
import { getTestData } from '@/utils/request';
import { loadTestData } from '@/api/test-data';

// Read settings from localStorage
function getSettings() {
  const settingsStr = localStorage.getItem("WeKnoRust_settings") ?? localStorage.getItem("WeKnora_settings");
  if (settingsStr) {
    try {
      const settings = JSON.parse(settingsStr);
      return settings;
    } catch (e) {
      console.error("Failed to parse settings:", e);
    }
  }
  return null;
}

interface StreamOptions {
  // HTTP method (default POST)
  method?: 'GET' | 'POST'
  // Request headers
  headers?: Record<string, string>
  // Request body to be auto-serialized
  body?: Record<string, any>
  // Streaming render interval (ms)
  chunkInterval?: number
}

export function useStream() {
  // Reactive state
  const output = ref('')              // Display content
  const isStreaming = ref(false)      // Streaming state
  const isLoading = ref(false)        // Initial loading flag
  const error = ref<string | null>(null)// Error message
  let controller = new AbortController()

  // Streaming render buffer
  let buffer: string[] = []
  let renderTimer: number | null = null

  // Start streaming request
  const startStream = async (params: { session_id: any; query: any; method: string; url: string }) => {
    // Reset state
    output.value = '';
    error.value = null;
    isStreaming.value = true;
    isLoading.value = true;

    // Get settings
    const settings = getSettings();
    let apiUrl = '';
    let apiKey = '';

    // Prefer settings if available
    if (settings && settings.endpoint && settings.apiKey) {
      apiUrl = settings.endpoint;
      apiKey = settings.apiKey;
    } else {
      // Otherwise load test data
      await loadTestData();
      const testData = getTestData();
      if (!testData) {
        error.value = "Test data not initialized; cannot start chat";
        stopStream();
        return;
      }
      apiUrl = import.meta.env.VITE_IS_DOCKER ? "" : "http://localhost:8080";
      apiKey = testData.tenant.api_key;
    }

    try {
      let url =
        params.method == "POST"
          ? `${apiUrl}${params.url}/${params.session_id}`
          : `${apiUrl}${params.url}/${params.session_id}?message_id=${params.query}`;
      await fetchEventSource(url, {
        method: params.method,
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": apiKey,
          "X-Request-ID": `${generateRandomString(12)}`,
        },
        body:
          params.method == "POST"
            ? JSON.stringify({ query: params.query })
            : null,
        signal: controller.signal,
        openWhenHidden: true,

        onopen: async (res) => {
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
          isLoading.value = false;
        },

        onmessage: (ev) => {
          buffer.push(JSON.parse(ev.data)); // push to buffer
          // Execute custom handler if provided
          if (chunkHandler) {
            chunkHandler(JSON.parse(ev.data));
          }
        },

        onerror: (err) => {
          throw new Error(`Streaming connection failed: ${err}`);
        },

        onclose: () => {
          stopStream();
        },
      });
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
      stopStream()
    }
  }

  let chunkHandler: ((data: any) => void) | null = null
  // Register chunk handler
  const onChunk = (handler: () => void) => {
    chunkHandler = handler
  }


  // Stop stream
  const stopStream = () => {
    controller.abort();
    controller = new AbortController(); // reset controller for next request
    isStreaming.value = false;
    isLoading.value = false;
  }

  // Auto cleanup on unmount
  onUnmounted(stopStream)

  return {
    output,          // Display content
    isStreaming,     // Whether streaming
    isLoading,       // Initial connection status
    error,
    onChunk,
    startStream,     // Start streaming
    stopStream       // Stop manually
  }
}