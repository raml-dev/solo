/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { writable } from "svelte/store";

export type NotificationType = "success" | "warning" | "error" | "info";

export interface Notification {
  id: string;
  type: NotificationType;
  message: string;
  detail?: string;
  persistent?: boolean;
}

function createNotificationStore() {
  const { subscribe, update } = writable<Notification[]>([]);

  function dismiss(id: string) {
    update((list) => list.filter((n) => n.id !== id));
  }

  function notify(
    type: NotificationType,
    message: string,
    detail?: string,
    persistent = false
  ): string {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
    update((list) => [...list, { id, type, message, detail, persistent }]);
    if (!persistent) {
      setTimeout(() => dismiss(id), 5000);
    }
    return id;
  }

  function dismissAll() {
    update(() => []);
  }

  function updateNotificationById(id: string, message: string, detail?: string) {
    update((list) => {
      const notification = list.find((n) => n.id === id);
      if (!notification) return list;
      notification.message = message;
      notification.detail = detail;
      return list;
    });
  }

  return {
    subscribe,
    notify,
    dismiss,
    dismissAll,
    success: (message: string, detail?: string) => notify("success", message, detail),
    error: (message: string, detail?: string, persistent = false) =>
      notify("error", message, detail, persistent),
    warning: (message: string, detail?: string) => notify("warning", message, detail),
    info: (message: string, detail?: string) => notify("info", message, detail),
    updateNotificationById
  };
}

export const notifications = createNotificationStore();
