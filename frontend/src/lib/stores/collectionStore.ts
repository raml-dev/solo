import { writable, derived } from "svelte/store";
import {
  CreateCollection,
  LoadCollections,
  LoadCollection,
  UpdateCollection,
  DeleteCollection,
  AddRequest,
  RemoveRequest,
  UpdateRequest
} from "../../../wailsjs/go/main/App";
import { collection } from "../../../wailsjs/go/models";
import { notifications } from "./notificationStore";

// Store state
interface CollectionState {
  collections: collection.Collection[];
  selectedCollectionName: string | null;
  selectedRequestId: string | null;
  loading: boolean;
}

const initialState: CollectionState = {
  collections: [],
  selectedCollectionName: null,
  selectedRequestId: null,
  loading: false
};

// Create the main store
function createCollectionStore() {
  const { subscribe, update } = writable<CollectionState>(initialState);

  return {
    subscribe,

    // Load all collections
    async loadCollections() {
      update((state) => ({ ...state, loading: true }));
      try {
        const collectionNames = await LoadCollections();
        if (!collectionNames || collectionNames.length === 0) {
          update((state) => ({ ...state, collections: [], loading: false }));
          return;
        }

        const collections: collection.Collection[] = [];
        for (const name of collectionNames) {
          const cleanName = name.replace(".json", "");
          try {
            const collection = await LoadCollection(cleanName);
            if (collection) collections.push(collection);
          } catch (err) {
            notifications.error(`Failed to load collection "${cleanName}"`, String(err));
          }
        }

        update((state) => ({ ...state, collections, loading: false }));
      } catch (err) {
        notifications.error("Failed to load collections", String(err), true);
        update((state) => ({ ...state, loading: false }));
      }
    },

    // Create a new collection
    async createCollection(name: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        await CreateCollection(name);
        await this.loadCollections();
      } catch (err) {
        notifications.error("Failed to create collection", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    // Delete a collection
    async deleteCollection(name: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        await DeleteCollection(name);
        update((state) => {
          const newState = {
            ...state,
            collections: state.collections.filter((c) => c.name !== name),
            loading: false
          };
          if (state.selectedCollectionName === name) {
            newState.selectedCollectionName = null;
            newState.selectedRequestId = null;
          }
          return newState;
        });
      } catch (err) {
        notifications.error("Failed to delete collection", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    // Rename a collection
    async renameCollection(currentName: string, newName: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        if (!currentName || !newName) throw new Error("Collection name is not specified");

        const existing = await LoadCollection(currentName);
        if (!existing) throw new Error(`Collection ${currentName} not found`);

        const updated = collection.Collection.createFrom({
          ...existing,
          name: newName,
          lastUpdateTimestamp: new Date().toISOString()
        });

        await UpdateCollection(updated);
        if (currentName !== newName) await DeleteCollection(currentName);
        await this.loadCollections();

        update((state) => ({
          ...state,
          selectedCollectionName:
            state.selectedCollectionName === currentName ? newName : state.selectedCollectionName
        }));
      } catch (err) {
        notifications.error("Failed to rename collection", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    // Add a request to a collection
    async addRequest(
      collectionName: string,
      request: Partial<collection.Request>
    ): Promise<collection.Request> {
      update((state) => ({ ...state }));
      try {
        const newRequestPayload = collection.Request.createFrom({
          id: "",
          name: request.name || "New Request",
          url: request.url || "",
          verb: request.verb || "GET",
          body: request.body || "",
          headers: request.headers || {},
          cookies: request.cookies || {},
          creationTimestamp: new Date().toISOString(),
          lastUpdateTimestamp: new Date().toISOString()
        });

        const newRequest = await AddRequest(collectionName, newRequestPayload);

        update((state) => {
          const updatedCollections = state.collections.map((c) => {
            if (c.name === collectionName) {
              return collection.Collection.createFrom({
                ...c,
                requests: [newRequest, ...c.requests],
                lastUpdateTimestamp: new Date().toISOString()
              });
            }
            return c;
          });
          return { ...state, collections: updatedCollections, selectedRequestId: newRequest.id };
        });

        return newRequest;
      } catch (err) {
        notifications.error("Failed to add request", String(err));
        throw err;
      }
    },

    // Update a request
    async updateRequest(collectionName: string, request: collection.Request) {
      try {
        await UpdateRequest(collectionName, request);
        update((state) => ({
          ...state,
          collections: state.collections.map((c) => {
            if (c.name !== collectionName) return c;
            return collection.Collection.createFrom({
              ...c,
              requests: c.requests.map((r) => (r.id === request.id ? request : r)),
              lastUpdateTimestamp: new Date().toISOString()
            });
          })
        }));
      } catch (err) {
        notifications.error("Failed to save request", String(err));
        throw err;
      }
    },

    // Remove a request
    async removeRequest(collectionName: string, requestId: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        await RemoveRequest(collectionName, requestId);
        update((state) => {
          const newState = {
            ...state,
            collections: state.collections.map((c) => {
              if (c.name === collectionName) {
                return collection.Collection.createFrom({
                  ...c,
                  requests: c.requests.filter((r) => r.id !== requestId)
                });
              }
              return c;
            }),
            loading: false
          };
          if (state.selectedRequestId === requestId) newState.selectedRequestId = null;
          return newState;
        });
      } catch (err) {
        notifications.error("Failed to remove request", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    // Select a collection
    selectCollection(name: string | null) {
      update((state) => ({ ...state, selectedCollectionName: name, selectedRequestId: null }));
    },

    // Select a request
    selectRequest(requestId: string | null) {
      update((state) => ({ ...state, selectedRequestId: requestId }));
    }
  };
}

export const collectionStore = createCollectionStore();

// Derived stores for convenient access
export const selectedCollection = derived(
  collectionStore,
  ($store) => $store.collections.find((c) => c.name === $store.selectedCollectionName) || null
);

export const selectedRequest = derived(
  [collectionStore, selectedCollection],
  ([$store, $collection]) => {
    if (!$collection || !$store.selectedRequestId) return null;
    return $collection.requests.find((r) => r.id === $store.selectedRequestId) || null;
  }
);
