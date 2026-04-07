export type LocalImportIcon = "upload" | "folder" | "document";

export interface LocalImportFormatOption<TFormat extends string = string> {
  key: TFormat;
  label: string;
  dropTitle: string;
  dropSubtitle: string;
  pickerButtonLabel: string;
  icon: LocalImportIcon;
}
