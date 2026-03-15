<script lang="ts">
  interface Props {
    variant?: BtnVariants;
    size?: BtnSize;
    disabled?: boolean;
    style?: string;
    square?: boolean;
    click: () => void;
    children?: import("svelte").Snippet;
  }

  let {
    variant = "primary",
    size = "sm",
    disabled = false,
    style = "",
    square = false,
    click,
    children
  }: Props = $props();

  type BtnVariants = "primary" | "secondary" | "tertiary" | "danger" | "success";
  type BtnSize = "sm" | "lg" | "small";

  const handleClick = () => click();
</script>

<button
  class="btn btn-{variant} btn-{size}"
  class:btn-square={square}
  {disabled}
  onclick={handleClick}
  {style}
  type="button">{@render children?.()}</button
>
