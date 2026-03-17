<script lang="ts">
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { environmentStore, selectedEnvironment } from "$src/lib/stores/environmentStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import {
    cancelHideTokenTooltip,
    forceHideTokenTooltip,
    tokenTooltipStore,
    tooltipMouseLeave
  } from "$src/lib/stores/tokenTooltipStore";
  import { environment } from "$wails/go/models";
  import { tick } from "svelte";

  let isEditing = $state(false);
  let editValue = $state("");
  let inputElement: HTMLInputElement | undefined = $state();

  let tooltipState = $derived($tokenTooltipStore);
  let tokenKey = $derived(tooltipState.tokenKey);
  let visible = $derived(tooltipState.visible);

  let sessionValue = $derived(tokenKey ? $sessionVarsStore[tokenKey] : undefined);
  let envValue = $derived($selectedEnvironment?.values[tokenKey]?.value ?? "");
  let hasSessionValue = $derived(sessionValue !== undefined);
  let hasEnvValue = $derived($selectedEnvironment?.values[tokenKey] !== undefined);
  let displayValue = $derived(hasSessionValue ? String(sessionValue) : envValue);
  let exists = $derived(hasSessionValue || hasEnvValue);
  let valueSource = $derived(hasSessionValue ? "session" : hasEnvValue ? "environment" : "none");

  $effect(() => {
    if (visible && !isEditing) {
      editValue = envValue;
    }
  });

  $effect(() => {
    if (!visible) {
      isEditing = false;
    }
  });

  async function handleEditClick() {
    isEditing = true;
    await tick();
    inputElement?.focus();
  }

  async function save() {
    if (!$selectedEnvironment) return;
    const processedValues = { ...$selectedEnvironment.values };
    processedValues[tokenKey] = new environment.ValueType({
      value: editValue,
      type: "string"
    });
    const envToSave = new environment.Environment({
      ...$selectedEnvironment,
      values: processedValues
    });
    await environmentStore.updateEnvironment(envToSave);
    isEditing = false;
  }

  async function clearSessionOverride() {
    if (!tokenKey) return;
    await sessionVarsStore.remove(tokenKey);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      save();
    } else if (e.key === "Escape") {
      isEditing = false;
      forceHideTokenTooltip();
    }
  }
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed z-50 w-max rounded-lg border border-neutral-200 bg-white p-3 shadow-lg dark:border-neutral-700 dark:bg-neutral-800"
    style="left: {tooltipState.x}px; top: {tooltipState.y}px;"
    onmouseenter={cancelHideTokenTooltip}
    onmouseleave={tooltipMouseLeave}
  >
    {#if isEditing}
      <div class="flex flex-col gap-2">
        <Input
          bind:elementRef={inputElement}
          type="text"
          size="sm"
          bind:value={editValue}
          onkeydown={handleKeydown}
        />
        <div class="flex items-center gap-2">
          <Button color="primary" size="xs" onclick={save}>Save</Button>
          <Button color="light" size="xs" onclick={() => (isEditing = false)}>Cancel</Button>
        </div>
      </div>
    {:else}
      <div class="flex items-center gap-2">
        <span
          class="font-mono text-sm {exists
            ? 'text-neutral-800 dark:text-neutral-100'
            : 'text-neutral-400 italic dark:text-neutral-500'}"
        >
          {exists ? displayValue : "Unresolved variable"}
        </span>
        {#if valueSource !== "none"}
          <Badge color={valueSource === "session" ? "yellow" : "blue"}>
            {valueSource}
          </Badge>
        {/if}
        <Button
          color="light"
          size="xs"
          onclick={handleEditClick}
          title="Edit environment value"
          aria-label="Edit environment value"
        >
          ✎
        </Button>
      </div>
      {#if valueSource === "session"}
        <div class="mt-2 flex flex-col gap-1">
          <p class="text-xs text-neutral-500 dark:text-neutral-400">
            Session override attivo: modificando qui cambi l'env salvato, ma il valore effettivo
            resta quello di sessione finché non la svuoti.
          </p>
          <Button color="light" size="xs" onclick={clearSessionOverride}>Use env value</Button>
        </div>
      {/if}
    {/if}
  </div>
{/if}
