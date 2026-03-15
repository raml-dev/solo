<script lang="ts">
  import type { ThemeSeeds } from "$src/lib/theme/themeModel";

  interface Props {
    seeds?: ThemeSeeds;
  }

  let { seeds }: Props = $props();

  let bg = $derived(seeds?.surface || "#f5f5f5");
  let sidebar = $derived(seeds?.neutral || "#e8e8e8");
  let content = $derived("#ffffff");
  let primary = $derived(seeds?.primary || "#3b82f6");
  let accent = $derived(seeds?.danger || "#ef4444");
  let text = $derived(seeds?.neutral || "#888");
  let border = $derived(seeds?.neutral || "rgba(0,0,0,0.1)");
</script>

<svg width="100%" height="100%" viewBox="0 0 240 130" xmlns="http://www.w3.org/2000/svg">
  <rect width="240" height="130" fill={bg} />

  <rect x="0" y="0" width="58" height="130" fill={sidebar} />
  <rect x="10" y="12" width="38" height="5" rx="2" fill={text} opacity=".4" />
  <rect x="10" y="24" width="30" height="4" rx="2" fill={text} opacity=".25" />
  <rect x="10" y="34" width="34" height="4" rx="2" fill={text} opacity=".25" />
  <rect x="10" y="44" width="28" height="4" rx="2" fill={text} opacity=".2" />

  <rect x="58" y="0" width="1" height="130" fill={border} opacity=".35" />

  <rect x="66" y="10" width="166" height="16" rx="3" fill={sidebar} opacity=".65" />
  <rect x="70" y="14" width="60" height="8" rx="2" fill={content} opacity=".8" />
  <rect x="136" y="14" width="40" height="8" rx="2" fill={content} opacity=".5" />
  <rect x="214" y="14" width="14" height="8" rx="2" fill={primary} />

  <rect x="66" y="32" width="166" height="1" fill={border} opacity=".35" />
  <rect x="66" y="40" width="100" height="7" rx="2" fill={text} opacity=".2" />
  <rect x="66" y="53" width="160" height="32" rx="3" fill={sidebar} opacity=".55" />
  <rect x="72" y="60" width="80" height="4" rx="2" fill={text} opacity=".25" />
  <rect x="72" y="70" width="60" height="4" rx="2" fill={text} opacity=".2" />
  <rect x="196" y="57" width="26" height="14" rx="3" fill={primary} />

  <rect x="66" y="91" width="36" height="5" rx="2" fill={accent} opacity=".6" />
  <rect x="106" y="91" width="50" height="5" rx="2" fill={text} opacity=".2" />
</svg>
