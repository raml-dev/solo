/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import {
  AddFolder,
  AddRequest,
  AddRequestToFolder,
  CreateCollection,
  DeleteCollection,
  GetRequests,
  LoadCollection,
  LoadCollections,
  RemoveFolder,
  RemoveRequest,
  UpdateCollection,
  UpdateFolder,
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

function insertRequestIntoList(
  requests: collection.Request[],
  request: collection.Request,
  targetRequestId?: string | null,
  position?: "before" | "after" | null
): collection.Request[] | null {
  if (!targetRequestId) {
    return [...requests, request];
  }

  const targetIndex = requests.findIndex((currentRequest) => currentRequest.id === targetRequestId);
  if (targetIndex === -1 || !position) {
    return null;
  }

  const nextRequests = [...requests];
  const insertAt = position === "before" ? targetIndex : targetIndex + 1;
  nextRequests.splice(insertAt, 0, request);
  return nextRequests;
}

function extractRequestFromFolders(
  folders: collection.Folder[],
  requestId: string
): {
  folders: collection.Folder[];
  extractedRequest: collection.Request | null;
} {
  let extractedRequest: collection.Request | null = null;

  const nextFolders = folders.map((folder) => {
    if (extractedRequest) {
      return folder;
    }

    const matchingRequest = (folder.requests || []).find(
      (currentRequest) => currentRequest.id === requestId
    );
    if (matchingRequest) {
      extractedRequest = matchingRequest;
      return collection.Folder.createFrom({
        ...folder,
        requests: (folder.requests || []).filter(
          (currentRequest) => currentRequest.id !== requestId
        ),
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });
    }

    const nestedResult = extractRequestFromFolders(folder.folders || [], requestId);
    if (!nestedResult.extractedRequest) {
      return folder;
    }

    extractedRequest = nestedResult.extractedRequest;
    return collection.Folder.createFrom({
      ...folder,
      folders: nestedResult.folders,
      lastUpdateTimestamp: new SvelteDate().toISOString()
    });
  });

  return { folders: nextFolders, extractedRequest };
}

function insertRequestIntoFolderTree(
  folders: collection.Folder[],
  targetFolderId: string,
  request: collection.Request,
  targetRequestId?: string | null,
  position?: "before" | "after" | null
): {
  folders: collection.Folder[];
  inserted: boolean;
} {
  let inserted = false;

  const nextFolders = folders.map((folder) => {
    if (inserted) {
      return folder;
    }

    if (folder.id === targetFolderId) {
      const nextRequests = insertRequestIntoList(
        folder.requests || [],
        request,
        targetRequestId,
        position
      );

      if (!nextRequests) {
        return folder;
      }

      inserted = true;
      return collection.Folder.createFrom({
        ...folder,
        requests: nextRequests,
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });
    }

    const nestedResult = insertRequestIntoFolderTree(
      folder.folders || [],
      targetFolderId,
      request,
      targetRequestId,
      position
    );

    if (!nestedResult.inserted) {
      return folder;
    }

    inserted = true;
    return collection.Folder.createFrom({
      ...folder,
      folders: nestedResult.folders,
      lastUpdateTimestamp: new SvelteDate().toISOString()
    });
  });

  return { folders: nextFolders, inserted };
}

function addFolderToTree(
  folders: collection.Folder[],
  parentFolderId: string | null,
  newFolder: collection.Folder
): collection.Folder[] {
  if (!parentFolderId) {
    return [...folders, newFolder];
  }

  return folders.map((folder) => {
    if (folder.id === parentFolderId) {
      return collection.Folder.createFrom({
        ...folder,
        folders: [...(folder.folders || []), newFolder],
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });
    }

    return collection.Folder.createFrom({
      ...folder,
      folders: addFolderToTree(folder.folders || [], parentFolderId, newFolder)
    });
  });
}

function updateFolderInTree(
  folders: collection.Folder[],
  updatedFolder: collection.Folder
): collection.Folder[] {
  return folders.map((folder) => {
    if (folder.id === updatedFolder.id) {
      return collection.Folder.createFrom({
        ...folder,
        ...updatedFolder
      });
    }

    return collection.Folder.createFrom({
      ...folder,
      folders: updateFolderInTree(folder.folders || [], updatedFolder)
    });
  });
}

function removeFolderFromTree(folders: collection.Folder[], folderId: string): collection.Folder[] {
  return folders
    .filter((folder) => folder.id !== folderId)
    .map((folder) =>
      collection.Folder.createFrom({
        ...folder,
        folders: removeFolderFromTree(folder.folders || [], folderId)
      })
    );
}

function addRequestToFolderInTree(
  folders: collection.Folder[],
  folderId: string,
  newRequest: collection.Request
): collection.Folder[] {
  return folders.map((folder) => {
    if (folder.id === folderId) {
      return collection.Folder.createFrom({
        ...folder,
        requests: [...(folder.requests || []), newRequest],
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });
    }

    return collection.Folder.createFrom({
      ...folder,
      folders: addRequestToFolderInTree(folder.folders || [], folderId, newRequest)
    });
  });
}

function updateRequestInFolders(
  folders: collection.Folder[],
  request: collection.Request
): collection.Folder[] {
  return folders.map((folder) =>
    collection.Folder.createFrom({
      ...folder,
      requests: (folder.requests || []).map((currentRequest) =>
        currentRequest.id === request.id ? request : currentRequest
      ),
      folders: updateRequestInFolders(folder.folders || [], request)
    })
  );
}

function removeRequestFromFolders(
  folders: collection.Folder[],
  requestId: string
): collection.Folder[] {
  return folders.map((folder) =>
    collection.Folder.createFrom({
      ...folder,
      requests: (folder.requests || []).filter((currentRequest) => currentRequest.id !== requestId),
      folders: removeRequestFromFolders(folder.folders || [], requestId)
    })
  );
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

  //Get Request of a collection
  async getRequestsFromCollection(name: string) {
    collectionStoreState.loading = true;
    try {
      return await GetRequests(name);
    } catch (err) {
      notifications.error("Failed to get request from collection", String(err));
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

  async updateCollectionVariables(
    collectionName: string,
    values: Record<string, { value: string; type: string }>
  ) {
    const targetCollection = collectionStoreState.collections.find(
      (currentCollection) => currentCollection.name === collectionName
    );

    if (!targetCollection) {
      throw new Error(`Collection ${collectionName} not found`);
    }

    const previousCollections = collectionStoreState.collections;
    const updatedCollection = collection.Collection.createFrom({
      ...targetCollection,
      variables: Object.fromEntries(
        Object.entries(values).map(([key, value]) => [key, new collection.ValueType(value)])
      ),
      lastUpdateTimestamp: new SvelteDate().toISOString()
    });

    collectionStoreState.collections = collectionStoreState.collections.map((currentCollection) =>
      currentCollection.name === collectionName ? updatedCollection : currentCollection
    );

    try {
      await UpdateCollection(updatedCollection);
    } catch (err) {
      collectionStoreState.collections = previousCollections;
      notifications.error("Failed to update collection variables", String(err));
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

  async addRequestToFolder(
    collectionName: string,
    folderId: string,
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

      const newRequest = await AddRequestToFolder(collectionName, folderId, newRequestPayload);

      collectionStoreState.collections = collectionStoreState.collections.map(
        (currentCollection) => {
          if (currentCollection.name !== collectionName) return currentCollection;

          return collection.Collection.createFrom({
            ...currentCollection,
            folders: addRequestToFolderInTree(
              currentCollection.folders || [],
              folderId,
              newRequest
            ),
            lastUpdateTimestamp: new SvelteDate().toISOString()
          });
        }
      );

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
          folders: updateRequestInFolders(c.folders || [], request),
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
            requests: c.requests.filter((r) => r.id !== requestId),
            folders: removeRequestFromFolders(c.folders || [], requestId)
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

  async addFolder(collectionName: string, parentFolderId: string | null, name: string) {
    try {
      const newFolderPayload = collection.Folder.createFrom({
        id: "",
        name,
        requests: [],
        folders: [],
        creationTimestamp: new SvelteDate().toISOString(),
        lastUpdateTimestamp: new SvelteDate().toISOString()
      });

      const newFolder = await AddFolder(collectionName, parentFolderId || "", newFolderPayload);

      collectionStoreState.collections = collectionStoreState.collections.map(
        (currentCollection) => {
          if (currentCollection.name !== collectionName) return currentCollection;

          return collection.Collection.createFrom({
            ...currentCollection,
            folders: addFolderToTree(currentCollection.folders || [], parentFolderId, newFolder),
            lastUpdateTimestamp: new SvelteDate().toISOString()
          });
        }
      );

      return newFolder;
    } catch (err) {
      notifications.error("Failed to create folder", String(err));
      throw err;
    }
  },

  async updateFolder(collectionName: string, folder: collection.Folder) {
    try {
      await UpdateFolder(collectionName, folder);

      collectionStoreState.collections = collectionStoreState.collections.map(
        (currentCollection) => {
          if (currentCollection.name !== collectionName) return currentCollection;

          return collection.Collection.createFrom({
            ...currentCollection,
            folders: updateFolderInTree(currentCollection.folders || [], folder),
            lastUpdateTimestamp: new SvelteDate().toISOString()
          });
        }
      );
    } catch (err) {
      notifications.error("Failed to update folder", String(err));
      throw err;
    }
  },

  async removeFolder(collectionName: string, folderId: string) {
    try {
      await RemoveFolder(collectionName, folderId);

      collectionStoreState.collections = collectionStoreState.collections.map(
        (currentCollection) => {
          if (currentCollection.name !== collectionName) return currentCollection;

          return collection.Collection.createFrom({
            ...currentCollection,
            folders: removeFolderFromTree(currentCollection.folders || [], folderId),
            lastUpdateTimestamp: new SvelteDate().toISOString()
          });
        }
      );
    } catch (err) {
      notifications.error("Failed to remove folder", String(err));
      throw err;
    }
  },

  async moveRequest(
    collectionName: string,
    requestId: string,
    sourceParentFolderId: string | null,
    targetParentFolderId: string | null,
    targetRequestId: string | null = null,
    position: "before" | "after" | null = null
  ) {
    const targetCollection = collectionStoreState.collections.find(
      (currentCollection) => currentCollection.name === collectionName
    );

    if (!targetCollection) {
      return;
    }

    if (sourceParentFolderId === targetParentFolderId && !targetRequestId) {
      return;
    }

    const previousCollections = collectionStoreState.collections;

    let extractedRequest: collection.Request;
    let nextRootRequests = [...(targetCollection.requests || [])];
    let nextFolders = [...(targetCollection.folders || [])];

    if (!sourceParentFolderId) {
      const sourceIndex = nextRootRequests.findIndex(
        (currentRequest) => currentRequest.id === requestId
      );
      if (sourceIndex === -1) {
        return;
      }

      const removedRequest = nextRootRequests[sourceIndex];
      if (!removedRequest) {
        return;
      }

      extractedRequest = removedRequest;
      nextRootRequests = nextRootRequests.filter(
        (currentRequest) => currentRequest.id !== requestId
      );
    } else {
      const extraction = extractRequestFromFolders(nextFolders, requestId);
      if (!extraction.extractedRequest) {
        return;
      }

      extractedRequest = extraction.extractedRequest;
      nextFolders = extraction.folders;
    }

    if (!targetParentFolderId) {
      const insertedRequests = insertRequestIntoList(
        nextRootRequests,
        extractedRequest,
        targetRequestId,
        position
      );

      if (!insertedRequests) {
        return;
      }

      nextRootRequests = insertedRequests;
    } else {
      const insertion = insertRequestIntoFolderTree(
        nextFolders,
        targetParentFolderId,
        extractedRequest,
        targetRequestId,
        position
      );

      if (!insertion.inserted) {
        return;
      }

      nextFolders = insertion.folders;
    }

    const updatedCollection = collection.Collection.createFrom({
      ...targetCollection,
      requests: nextRootRequests,
      folders: nextFolders,
      lastUpdateTimestamp: new SvelteDate().toISOString()
    });

    collectionStoreState.collections = collectionStoreState.collections.map((currentCollection) =>
      currentCollection.name === collectionName ? updatedCollection : currentCollection
    );

    try {
      await UpdateCollection(updatedCollection);
    } catch (err) {
      collectionStoreState.collections = previousCollections;
      notifications.error("Failed to move request", String(err));
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
