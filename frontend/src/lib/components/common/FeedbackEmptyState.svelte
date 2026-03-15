<script lang="ts">
  import Alert from "flowbite-svelte/Alert.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import type { Snippet } from "svelte";

  type FeedbackVariant = "neutral" | "info" | "warning" | "danger";

  interface Props {
    variant?: FeedbackVariant;
    title: string;
    detail?: string;
    compact?: boolean;
    iconSnippet?: Snippet;
    actions?: Snippet;
  }

  let { variant = "neutral", title, detail, compact = false, iconSnippet, actions }: Props =
    $props();

  const useAlert = $derived(variant === "warning" || variant === "danger");

  const alertColor = $derived(
    variant === "danger" ? "red" : variant === "warning" ? "yellow" : "blue"
  );

  const cardClasses = $derived(
    compact ? "w-full px-3 py-4 text-center" : "w-full px-5 py-8 text-center"
  );
</script>

{#if useAlert}
  <Alert color={alertColor} class={compact ? "p-3" : "p-5"}>
    <div class="flex flex-col gap-2 text-center sm:text-left">
      {#if iconSnippet}
        <div class="mx-auto sm:mx-0">
          {@render iconSnippet()}
        </div>
      {/if}
      <p class="font-semibold">{title}</p>
      {#if detail}
        <p class="text-sm opacity-90">{detail}</p>
      {/if}
      {#if actions}
        <div class="mt-2 flex flex-wrap items-center justify-center gap-2 sm:justify-start">
          {@render actions()}
        </div>
      {/if}
    </div>
  </Alert>
{:else}
  <Card class={cardClasses}>
    <div class="flex flex-col items-center gap-2">
      {#if iconSnippet}
        <div class="text-neutral-500 dark:text-neutral-300">
          {@render iconSnippet()}
        </div>
      {/if}
      <p class="font-medium text-neutral-900 dark:text-neutral-100">{title}</p>
      {#if detail}
        <p class="text-sm text-neutral-600 dark:text-neutral-300">{detail}</p>
      {/if}
      {#if actions}
        <div class="mt-2 flex flex-wrap items-center justify-center gap-2">
          {@render actions()}
        </div>
      {/if}
    </div>
  </Card>
{/if}
