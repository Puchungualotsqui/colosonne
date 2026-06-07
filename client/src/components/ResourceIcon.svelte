<script lang="ts">
    export let resource: "wood" | "stone" | "grain";
    export let amount: number | string = "";
    export let size: "sm" | "md" = "sm";
    export let pulse = false;

    function icon() {
        switch (resource) {
            case "wood":
                return "♣";
            case "stone":
                return "▲";
            case "grain":
                return "◆";
        }
    }

    function title() {
        switch (resource) {
            case "wood":
                return "Wood";
            case "stone":
                return "Stone";
            case "grain":
                return "Grain";
        }
    }

    function colorClass() {
        switch (resource) {
            case "wood":
                return "border-[#2f6546] bg-[#5b9368] text-[#142833]";
            case "stone":
                return "border-[#656b73] bg-[#a8adb2] text-[#142833]";
            case "grain":
                return "border-[#9b7034] bg-[#d9b56a] text-[#142833]";
        }
    }

    $: sizeClass =
        size === "md"
            ? "h-10 min-w-16 gap-2 px-3 text-sm"
            : "h-8 min-w-12 gap-1.5 px-2 text-xs";

    $: iconClass = size === "md" ? "text-xl" : "text-base";
</script>

<div
    class={[
        "inline-flex items-center justify-center rounded-xl border-2 font-black shadow-sm transition",
        sizeClass,
        colorClass(),
        pulse ? "resource-pulse" : "",
    ].join(" ")}
    title={title()}
>
    <span class={iconClass}>{icon()}</span>

    {#if amount !== ""}
        <span>{amount}</span>
    {/if}
</div>

<style>
    .resource-pulse {
        animation: resource-pop 720ms ease-out;
    }

    @keyframes resource-pop {
        0% {
            transform: scale(1);
            filter: brightness(1);
        }

        35% {
            transform: scale(1.16);
            filter: brightness(1.22);
        }

        100% {
            transform: scale(1);
            filter: brightness(1);
        }
    }
</style>
