import { writable } from "svelte/store";

export interface HistoryRequest {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: string;
}

export interface HistoryResponse {
  status: number;
  time: number;
  headers: Record<string, string>;
  body: string;
}

export interface HistoryEntry {
  id: string;
  timestamp: string;
  collectionName: string | null;
  requestName: string | null;
  request: HistoryRequest;
  response: HistoryResponse | null;
  error: string | null;
}

const MAX_ENTRIES = 500;

function createHistoryStore() {
  const { subscribe, update } = writable<HistoryEntry[]>([]);

  return {
    subscribe,

    push(entry: Omit<HistoryEntry, "id" | "timestamp">) {
      const newEntry: HistoryEntry = {
        ...entry,
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        timestamp: new Date().toISOString()
      };
      update((list) => {
        const next = [newEntry, ...list];
        return next.length > MAX_ENTRIES ? next.slice(0, MAX_ENTRIES) : next;
      });
    },

    clear() {
      update(() => []);
    },

    exportHAR(entries: HistoryEntry[]): string {
      const harEntries = entries.map((e) => ({
        startedDateTime: e.timestamp,
        time: e.response?.time ?? 0,
        request: {
          method: e.request.method,
          url: e.request.url,
          httpVersion: "HTTP/1.1",
          headers: Object.entries(e.request.headers).map(([name, value]) => ({ name, value })),
          queryString: [],
          cookies: [],
          headersSize: -1,
          bodySize: e.request.body?.length ?? 0,
          postData: e.request.body
            ? { mimeType: "application/json", text: e.request.body }
            : undefined
        },
        response: e.response
          ? {
              status: e.response.status,
              statusText: String(e.response.status),
              httpVersion: "HTTP/1.1",
              headers: Object.entries(e.response.headers).map(([name, value]) => ({ name, value })),
              cookies: [],
              content: {
                size: e.response.body?.length ?? 0,
                mimeType: e.response.headers["content-type"] ?? "text/plain",
                text: e.response.body
              },
              redirectURL: "",
              headersSize: -1,
              bodySize: e.response.body?.length ?? 0
            }
          : {
              status: 0,
              statusText: e.error ?? "Error",
              httpVersion: "HTTP/1.1",
              headers: [],
              cookies: [],
              content: { size: 0, mimeType: "text/plain", text: e.error ?? "" },
              redirectURL: "",
              headersSize: -1,
              bodySize: 0
            },
        cache: {},
        timings: { send: 0, wait: e.response?.time ?? 0, receive: 0 }
      }));

      return JSON.stringify(
        {
          log: {
            version: "1.2",
            creator: { name: "solo", version: "1.0" },
            entries: harEntries
          }
        },
        null,
        2
      );
    }
  };
}

export const historyStore = createHistoryStore();
