<script lang="ts">
    export let variant: "dark" | "board" | "innerBoard" = "dark";
    export let padding: "none" | "sm" | "md" | "lg" = "md";
    export let className = "";

    $: baseClass =
        variant === "board"
            ? "rounded-[34px] bg-[#caa66d] shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
            : variant === "innerBoard"
              ? "rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] text-[#142833] shadow-inner"
              : "rounded-3xl bg-[#23444c] shadow-md ring-1 ring-[#f8efe0]/10";

    $: paddingClass =
        padding === "none"
            ? ""
            : padding === "sm"
              ? "p-3"
              : padding === "lg"
                ? "p-6"
                : "p-5";
</script>

<section class={[baseClass, paddingClass, className].join(" ")}>
    <slot />
</section>
