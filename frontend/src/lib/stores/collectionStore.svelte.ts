/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import {
  AddRequest,
  CreateCollection,
  DeleteCollection,
  LoadCollection,
  LoadCollections,
  RemoveRequest,
  UpdateCollection,
  UpdateRequest
} from "$wails/go/main/App";
import { collection } from "$wails/go/models";
import { SvelteDate } from "svelte/reactivity";

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

function moveRequestById(
  requests: collection.Request[],
  sourceRequestId: string,
  targetRequestId: string,
  position: "before" | "after"
): collection.Request[] | null {
  if (!sourceRequestId || !targetRequestId || sourceRequestId === targetRequestId) {
    return null;
  }

  const sourceIndex = requests.findIndex((request) => request.id === sourceRequestId);
  const targetIndex = requests.findIndex((request) => request.id === targetRequestId);

  if (sourceIndex === -1 || targetIndex === -1) {
    return null;
  }

  const reorderedRequests = [...requests];
  const [sourceRequest] = reorderedRequests.splice(sourceIndex, 1);

  if (!sourceRequest) {
    return null;
  }

  const nextTargetIndex = reorderedRequests.findIndex((request) => request.id === targetRequestId);
  if (nextTargetIndex === -1) {
    return null;
  }

  const insertAt = position === "before" ? nextTargetIndex : nextTargetIndex + 1;
  reorderedRequests.splice(insertAt, 0, sourceRequest);

  const didChangeOrder = reorderedRequests.some(
    (request, index) => request.id !== requests[index]?.id
  );
  return didChangeOrder ? reorderedRequests : null;
}

export const collectionStoreState = $state<CollectionState>({ ...initialState });

export const collectionStore = {
  // Load all collections
  async loadCollections() {
    collectionStoreState.loading = true;
    try {
      const collectionNames = await LoadCollections();
      if (!collectionNames || collectionNames.length === 0) {
        collectionStoreState.collections = [];
        collectionStoreState.loading = false;
        return;
      }

      const collectionsLoaded: collection.Collection[] = [];
      for (const name of collectionNames) {
        const cleanName = name.replace(".json", "");
        try {
          const coll = await LoadCollection(cleanName);
          if (coll) collectionsLoaded.push(coll);
        } catch (err) {
          notifications.error(`Failed to load collection "${cleanName}"`, String(err));
        }
      }

      collectionStoreState.collections = collectionsLoaded;
      collectionStoreState.loading = false;
    } catch (err) {
      notifications.error("Failed to load collections", String(err), true);
      collectionStoreState.loading = false;
    }
  },

  // Create a new collection
  async createCollection(name: string) {
    collectionStoreState.loading = true;
    try {
      await CreateCollection(name);
      await this.loadCollections();
    } catch (err) {
      notifications.error("Failed to create collection", String(err));
      collectionStoreState.loading = false;
      throw err;
    }
  },

  // Delete a collection
  async deleteCollection(name: string) {
    collectionStoreState.loading = true;
    try {
      await DeleteCollection(name);
      const newCollections = collectionStoreState.collections.filter((c) => c.name !== name);
      collectionStoreState.collections = newCollections;
      if (collectionStoreState.selectedCollectionName === name) {
        collectionStoreState.selectedCollectionName = null;
        collectionStoreState.selectedRequestId = null;
      }
      collectionStoreState.loading = false;
    } catch (err) {
      notifications.error("Failed to delete collection", String(err));
      collectionStoreState.loading = false;
      throw err;
    }
  },

  // Rename a collection
  async renameCollection(currentName: string, newName: string) {
    collectionStoreState.loading = true;
    try {
      if (!currentName || !newName) throw new Error("Collection name is not specified");

      const existing = await LoadCollection(currentName);
      if (!existing) throw new Error(`Collection ${currentName} not found`);

      const updated = collection.Collection.createFrom({
        ...existing,
        name: newName,
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });

      await UpdateCollection(updated);
      if (currentName !== newName) await DeleteCollection(currentName);
      await this.loadCollections();

      if (collectionStoreState.selectedCollectionName === currentName) {
        collectionStoreState.selectedCollectionName = newName;
      }
      collectionStoreState.loading = false;
    } catch (err) {
      notifications.error("Failed to rename collection", String(err));
      collectionStoreState.loading = false;
      throw err;
    }
  },

  // Add a request to a collection
  async addRequest(
    collectionName: string,
    request: Partial<collection.Request>
  ): Promise<collection.Request> {
    try {
      const newRequestPayload = collection.Request.createFrom({
        id: "",
        name: request.name || "New Request",
        url: request.url || "",
        verb: request.verb || "GET",
        body: request.body || "",
        headers: request.headers || {},
        cookies: request.cookies || {},
        bodyType: request.bodyType || undefined,
        auth: request.auth || undefined,
        settings: request.settings || undefined,
        preRequestScript: request.preRequestScript || undefined,
        postResponseScript: request.postResponseScript || undefined,
        creationTimestamp: new SvelteDate().toISOString(),
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });

      const newRequest = await AddRequest(collectionName, newRequestPayload);

      const updatedCollections = collectionStoreState.collections.map((c) => {
        if (c.name === collectionName) {
          return collection.Collection.createFrom({
            ...c,
            requests: [...c.requests, newRequest],
            lastUpdateTimestamp: new SvelteDate().toISOString()
          });
        }
        return c;
      });

      collectionStoreState.collections = updatedCollections;
      collectionStoreState.selectedRequestId = newRequest.id;

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
      collectionStoreState.collections = collectionStoreState.collections.map((c) => {
        if (c.name !== collectionName) return c;
        return collection.Collection.createFrom({
          ...c,
          requests: c.requests.map((r) => (r.id === request.id ? request : r)),
          lastUpdateTimestamp: new SvelteDate().toISOString()
        });
      });
    } catch (err) {
      notifications.error("Failed to save request", String(err));
      throw err;
    }
  },

  // Remove a request
  async removeRequest(collectionName: string, requestId: string) {
    collectionStoreState.loading = true;
    try {
      await RemoveRequest(collectionName, requestId);
      const newCollections = collectionStoreState.collections.map((c) => {
        if (c.name === collectionName) {
          return collection.Collection.createFrom({
            ...c,
            requests: c.requests.filter((r) => r.id !== requestId)
          });
        }
        return c;
      });
      collectionStoreState.collections = newCollections;
      collectionStoreState.loading = false;
      if (collectionStoreState.selectedRequestId === requestId)
        collectionStoreState.selectedRequestId = null;
    } catch (err) {
      notifications.error("Failed to remove request", String(err));
      collectionStoreState.loading = false;
      throw err;
    }
  },

  // Reorder requests inside a collection
  async reorderRequests(
    collectionName: string,
    sourceRequestId: string,
    targetRequestId: string,
    position: "before" | "after"
  ) {
    const targetCollection = collectionStoreState.collections.find(
      (currentCollection) => currentCollection.name === collectionName
    );

    if (!targetCollection) {
      return;
    }

    const reorderedRequests = moveRequestById(
      targetCollection.requests || [],
      sourceRequestId,
      targetRequestId,
      position
    );

    if (!reorderedRequests) {
      return;
    }

    const previousCollections = collectionStoreState.collections;
    const updatedCollection = collection.Collection.createFrom({
      ...targetCollection,
      requests: reorderedRequests,
      lastUpdateTimestamp: new SvelteDate().toISOString()
    });

    collectionStoreState.collections = collectionStoreState.collections.map((currentCollection) =>
      currentCollection.name === collectionName ? updatedCollection : currentCollection
    );

    try {
      await UpdateCollection(updatedCollection);
    } catch (err) {
      collectionStoreState.collections = previousCollections;
      notifications.error("Failed to reorder requests", String(err));
      throw err;
    }
  },

  // Select a collection
  selectCollection(name: string | null) {
    collectionStoreState.selectedCollectionName = name;
    collectionStoreState.selectedRequestId = null;
  },

  // Select a request
  selectRequest(requestId: string | null) {
    collectionStoreState.selectedRequestId = requestId;
  }
};
