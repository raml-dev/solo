<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ExportModal, { type ExportOption } from "$src/lib/components/common/ExportModal.svelte";

  const EXPORT_OPTIONS: ExportOption[] = [
    {
      id: "solo",
      label: "Solo (JSON)",
      description: "Native Solo format, preserves all collection data"
    },
    {
      id: "openapi",
      label: "OpenAPI 3.1 (YAML)",
      description: "Standard OpenAPI format, compatible with other tools"
    }
  ];

  interface Props {
    open: boolean;
    onExport: (format: "solo" | "openapi") => void;
    onClose: () => void;
  }

  let { open = $bindable(), onExport, onClose }: Props = $props();
</script>

<ExportModal
  bind:open
  title="Export Collection"
  options={EXPORT_OPTIONS}
  onExport={(id) => onExport(id as "solo" | "openapi")}
  {onClose}
/>
