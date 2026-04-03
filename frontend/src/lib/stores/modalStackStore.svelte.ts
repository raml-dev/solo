/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { toStore } from "svelte/store";

interface ModalStackState {
  openOrder: string[];
  openById: Record<string, boolean>;
}

interface ModalHandle {
  id: string;
  get open(): boolean;
  set open(_value: boolean);
}

export const modalStackState: ModalStackState = $state({
  openOrder: [],
  openById: {}
});

function createModalId(scope = "modal"): string {
  const id = `${scope}-${Math.random().toString(36).slice(2)}`;
  modalStackState.openById[id] = false;
  return id;
}

function ensureModal(id: string) {
  if (!(id in modalStackState.openById)) {
    modalStackState.openById[id] = false;
  }
}

function setOpen(id: string, value: boolean) {
  ensureModal(id);

  const current = modalStackState.openById[id] === true;
  if (current === value) return;

  modalStackState.openById[id] = value;

  const withoutId = modalStackState.openOrder.filter((item) => item !== id);
  modalStackState.openOrder = value ? [...withoutId, id] : withoutId;
}

function open(id: string) {
  setOpen(id, true);
}

function close(id: string) {
  setOpen(id, false);
}

function isOpen(id: string): boolean {
  return modalStackState.openById[id] === true;
}

function destroyModal(id: string) {
  const nextOpenById = { ...modalStackState.openById };
  delete nextOpenById[id];
  modalStackState.openById = nextOpenById;
  modalStackState.openOrder = modalStackState.openOrder.filter((item) => item !== id);
}

function createModal(scope = "modal"): ModalHandle {
  const id = createModalId(scope);
  return {
    id,
    get open() {
      return isOpen(id);
    },
    set open(value: boolean) {
      setOpen(id, value);
    }
  };
}

export const topModalId = toStore(
  () => modalStackState.openOrder[modalStackState.openOrder.length - 1] ?? null
);
export const hasOpenModals = toStore(() => modalStackState.openOrder.length > 0);

export const modalStack = {
  createModalId,
  createModal,
  destroyModal,
  setOpen,
  open,
  close,
  isOpen
};
