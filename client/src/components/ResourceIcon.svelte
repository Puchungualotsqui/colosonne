<script lang="ts">
    export let resource: "wood" | "stone" | "grain" | "relic" = "wood";
    export let amount: number | string = 0;
    export let size: "sm" | "md" = "sm";
    export let pulse = false;

    function icon(resource: string) {
        switch (resource) {
            case "wood":
                return "♣";
            case "stone":
                return "▲";
            case "grain":
                return "◆";
            case "relic":
                return "✧";
            default:
                return "?";
        }
    }

    function label(resource: string) {
        switch (resource) {
            case "wood":
                return "Wood";
            case "stone":
                return "Stone";
            case "grain":
                return "Grain";
            case "relic":
                return "Relic";
            default:
                return "";
        }
    }

    function colorClass(resource: string) {
        switch (resource) {
            case "wood":
                return "border-[#2f6546] bg-[#5b9368] text-[#142833]";
            case "stone":
                return "border-[#656b73] bg-[#a8adb2] text-[#142833]";
            case "grain":
                return "border-[#9b7034] bg-[#d9b56a] text-[#142833]";
            case "relic":
                return "border-[#6d4c9b] bg-[#9b79c9] text-[#142833]";
            default:
                return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
        }
    }

    $: sizeClass =
        size === "md"
            ? "h-10 min-w-14 px-3 text-sm"
            : "h-8 min-w-12 px-2 text-xs";
</script>

<div
    class={[
        "inline-flex items-center justify-center gap-1.5 rounded-xl border-2 font-black shadow-sm",
        sizeClass,
        colorClass(resource),
        pulse ? "resource-pulse" : "",
    ].join(" ")}
    title={label(resource)}
>
    <span>{icon(resource)}</span>
    <span>{amount}</span>
</div>

<style>
    .resource-pulse {
        animation: resource-pulse 700ms ease-out;
    }

    @keyframes resource-pulse {
        0% {
            transform: scale(1);
        }

        35% {
            transform: scale(1.14);
        }

        100% {
            transform: scale(1);
        }
    }
</style>
