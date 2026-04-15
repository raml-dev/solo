/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

export type LocalImportIcon = "upload" | "folder" | "document";

export interface LocalImportFormatOption<TFormat extends string = string> {
  key: TFormat;
  label: string;
  dropTitle: string;
  dropSubtitle: string;
  pickerButtonLabel: string;
  icon: LocalImportIcon;
}
