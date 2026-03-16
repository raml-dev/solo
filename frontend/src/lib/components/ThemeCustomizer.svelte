<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { SaveCustomTheme } from "$wails/go/main/App";
  import { theme } from "$wails/go/models";

  interface Props {
    baseTheme?: theme.Theme | null;
    saved: (theme: theme.Theme) => void;
    close: () => void;
  }

  let { baseTheme = null, saved, close }: Props = $props();

  let themeName = $state("");
  let seeds = $state({
    primary: "#0ea5e9",
    success: "#10b981",
    warning: "#f59e0b",
    danger: "#ef4444",
    neutral: "#52525b",
    surface: ""
  });

  $effect(() => {
    if (baseTheme?.config?.seeds) {
      themeName = `${baseTheme.label}-custom`;
      seeds = {
        primary: baseTheme.config.seeds.primary,
        success: baseTheme.config.seeds.success,
        warning: baseTheme.config.seeds.warning,
        danger: baseTheme.config.seeds.danger,
        neutral: baseTheme.config.seeds.neutral,
        surface: baseTheme.config.seeds.surface || ""
      };
    }
  });

  function toId(name: string) {
    return `custom-${name
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "")}`;
  }

  async function saveTheme() {
    if (!themeName.trim()) {
      notifications.warning("Please enter a theme name");
      return;
    }

    const newTheme = new theme.Theme({
      id: toId(themeName),
      label: themeName,
      config: {
        type: "custom",
        mode: "system",
        seeds: {
          primary: seeds.primary,
          success: seeds.success,
          warning: seeds.warning,
          danger: seeds.danger,
          neutral: seeds.neutral,
          ...(seeds.surface ? { surface: seeds.surface } : {})
        }
      }
    });

    try {
      await SaveCustomTheme(newTheme);
      await configurationStore.changeTheme(newTheme.id);
      saved(newTheme);
      close();
    } catch (error) {
      notifications.error("Failed to save theme", String(error));
    }
  }
</script>

<div class="customizer">
  <div class="customizer-header">
    <h3 class="text-lg font-semibold">Customize Theme</h3>
    <Button color="primary" onclick={close} aria-label="Close theme customizer">×</Button>
  </div>

  <div class="customizer-content">
    <div class="form-group">
      <label for="theme-name" class="label">Theme Name</label>
      <input
        id="theme-name"
        type="text"
        class="input"
        bind:value={themeName}
        placeholder="My Custom Theme"
      />
    </div>

    <div class="color-grid">
      <label><span>Primary</span><input type="color" bind:value={seeds.primary} /></label>
      <label><span>Success</span><input type="color" bind:value={seeds.success} /></label>
      <label><span>Warning</span><input type="color" bind:value={seeds.warning} /></label>
      <label><span>Danger</span><input type="color" bind:value={seeds.danger} /></label>
      <label><span>Neutral</span><input type="color" bind:value={seeds.neutral} /></label>
      <label><span>Surface (optional)</span><input type="color" bind:value={seeds.surface} /></label
      >
    </div>
  </div>

  <div class="customizer-footer">
    <Button color="light" onclick={close}>Cancel</Button>
    <Button color="primary" onclick={saveTheme}>Save Theme</Button>
  </div>
</div>
