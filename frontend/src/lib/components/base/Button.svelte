<script lang="ts">
  export let variant: BtnVariants = "primary";
  export let size: BtnSize = "sm";
  export let disabled = false;
  export let style: string = "";
  export let square = false;

  export let click: () => void;

  type BtnVariants = "primary" | "secondary" | "tertiary" | "danger" | "success";
  type BtnSize = "sm" | "lg" | "small";

  const handleClick = () => click();
</script>

<button class="btn btn-{variant} btn-{size}" class:btn-square={square} {disabled} on:click={handleClick} {style} type="button"
  ><slot></slot></button
>

<style>
  /* Buttons */
  .btn {
    padding: var(--space-sm) var(--space-md);
    border: none;
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: all var(--transition-fast);
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
    font-family: var(--font-sans);
    align-items: center;
    justify-content: center;
  }

  .btn-primary {
    background: var(--primary);
    color: var(--bg-primary);
  }

  .btn-primary:hover {
    background: var(--primary-dark);
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text);
    border: 1px solid var(--border);
  }

  .btn-secondary:hover {
    background: var(--border);
  }

  .btn-danger {
    background: var(--danger);
    color: var(--bg-primary);
  }

  .btn-danger:hover {
    opacity: 0.9;
  }

  .btn-success {
    background: var(--success);
    color: var(--bg-primary);
  }

  .btn-success:hover {
    opacity: 0.9;
  }

  .btn-sm,
  .btn-small {
    padding: var(--space-xs) var(--space-sm);
    font-size: var(--font-size-xs);
  }

  .btn-lg {
    padding: var(--space-md) var(--space-lg);
    font-size: var(--font-size-base);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-square {
    border-radius: 0;
  }
</style>
