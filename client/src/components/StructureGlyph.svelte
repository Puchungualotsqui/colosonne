<script lang="ts" context="module">
    export type StructureGlyphKind =
        | "none"
        | "outpost"
        | "city"
        | "settlement"
        | "bridge"
        | "watchtower"
        | "road";
</script>

<script lang="ts">
    import { Structure } from "../lib/types";

    export let structure: Structure | StructureGlyphKind = Structure.None;
    export let owner: 0 | 1 | 2 | "blue" | "red" | "neutral" = "neutral";
    export let size: "sm" | "md" | "lg" = "md";
    export let boxed = false;
    export let title = "";

    $: kind = normalizeStructure(structure);

    $: sizeClass =
        size === "lg" ? "h-16 w-16" : size === "sm" ? "h-9 w-9" : "h-12 w-12";

    $: glyphScale =
        size === "lg"
            ? "scale-[1.15]"
            : size === "sm"
              ? "scale-[0.78]"
              : "scale-100";

    $: ownerClass = boxed ? boxedOwnerClass(owner) : "";

    function normalizeStructure(value: Structure | StructureGlyphKind) {
        if (typeof value === "string") return value;

        switch (value) {
            case Structure.Outpost:
                return "outpost";
            case Structure.City:
                return "city";
            case Structure.Settlement:
                return "settlement";
            case Structure.Bridge:
                return "bridge";
            case Structure.Watchtower:
                return "watchtower";
            default:
                return "none";
        }
    }

    function boxedOwnerClass(value: typeof owner) {
        if (value === 1 || value === "blue") {
            return "bg-[#1d4e89] text-white ring-[#f8efe0]/35";
        }

        if (value === 2 || value === "red") {
            return "bg-[#b94b3f] text-white ring-[#f8efe0]/35";
        }

        return "bg-[#f8efe0]/70 text-[#142833] ring-black/10";
    }
</script>

<div
    class={[
        "structure-glyph relative grid place-items-center",
        sizeClass,
        boxed ? "rounded-2xl shadow-sm ring-2" : "",
        ownerClass,
    ].join(" ")}
    {title}
    aria-label={title || kind}
>
    <div class={["glyph-inner relative", glyphScale].join(" ")}>
        {#if kind === "outpost"}
            <div class="outpost-roof"></div>
            <div class="outpost-body"></div>
            <div class="outpost-flag"></div>
        {:else if kind === "city"}
            <div class="city-block city-block-a"></div>
            <div class="city-block city-block-b"></div>
            <div class="city-block city-block-c"></div>
        {:else if kind === "settlement"}
            <div class="settlement-diamond"></div>
            <div class="settlement-core"></div>
        {:else if kind === "bridge"}
            <div class="bridge-deck"></div>
            <div class="bridge-arch bridge-arch-a"></div>
            <div class="bridge-arch bridge-arch-b"></div>
        {:else if kind === "watchtower"}
            <div class="tower-top"></div>
            <div class="tower-body"></div>
            <div class="tower-legs"></div>
        {:else if kind === "road"}
            <div class="road-line"></div>
            <div class="road-node road-node-a"></div>
            <div class="road-node road-node-b"></div>
        {:else}
            <div class="none-dot"></div>
        {/if}
    </div>
</div>

<style>
    .structure-glyph {
        line-height: 1;
        color: currentColor;
    }

    .glyph-inner {
        width: 42px;
        height: 42px;
    }

    .glyph-inner * {
        position: absolute;
        box-sizing: border-box;
    }

    .outpost-roof {
        left: 8px;
        top: 8px;
        width: 26px;
        height: 16px;
        background: currentColor;
        clip-path: polygon(50% 0%, 100% 72%, 0% 72%);
    }

    .outpost-body {
        left: 12px;
        top: 20px;
        width: 18px;
        height: 15px;
        border-radius: 4px 4px 2px 2px;
        background: currentColor;
    }

    .outpost-flag {
        left: 26px;
        top: 5px;
        width: 10px;
        height: 8px;
        background: currentColor;
        clip-path: polygon(0% 0%, 100% 18%, 100% 82%, 0% 100%);
    }

    .city-block {
        bottom: 7px;
        background: currentColor;
        border-radius: 3px 3px 1px 1px;
    }

    .city-block-a {
        left: 7px;
        width: 10px;
        height: 20px;
    }

    .city-block-b {
        left: 17px;
        width: 12px;
        height: 28px;
    }

    .city-block-c {
        left: 29px;
        width: 8px;
        height: 16px;
    }

    .settlement-diamond {
        left: 8px;
        top: 8px;
        width: 26px;
        height: 26px;
        background: currentColor;
        transform: rotate(45deg);
        border-radius: 4px;
    }

    .settlement-core {
        left: 17px;
        top: 17px;
        width: 8px;
        height: 8px;
        background: rgba(20, 40, 51, 0.75);
        transform: rotate(45deg);
        border-radius: 2px;
    }

    .bridge-deck {
        left: 6px;
        top: 15px;
        width: 30px;
        height: 6px;
        border-radius: 999px;
        background: currentColor;
    }

    .bridge-arch {
        top: 18px;
        width: 13px;
        height: 14px;
        border: 4px solid currentColor;
        border-top: 0;
        border-radius: 0 0 999px 999px;
    }

    .bridge-arch-a {
        left: 8px;
    }

    .bridge-arch-b {
        left: 21px;
    }

    .tower-top {
        left: 10px;
        top: 5px;
        width: 22px;
        height: 9px;
        border-radius: 3px;
        background: currentColor;
    }

    .tower-body {
        left: 15px;
        top: 14px;
        width: 12px;
        height: 18px;
        border-radius: 2px;
        background: currentColor;
    }

    .tower-legs {
        left: 10px;
        top: 29px;
        width: 22px;
        height: 10px;
        border-left: 5px solid currentColor;
        border-right: 5px solid currentColor;
        transform: perspective(20px) rotateX(12deg);
    }

    .road-line {
        left: 5px;
        top: 18px;
        width: 32px;
        height: 7px;
        border-radius: 999px;
        background: currentColor;
        transform: rotate(-24deg);
    }

    .road-node {
        width: 9px;
        height: 9px;
        border-radius: 999px;
        background: currentColor;
    }

    .road-node-a {
        left: 5px;
        top: 24px;
    }

    .road-node-b {
        right: 5px;
        top: 10px;
    }

    .none-dot {
        left: 17px;
        top: 17px;
        width: 8px;
        height: 8px;
        border-radius: 999px;
        background: currentColor;
        opacity: 0.5;
    }
</style>
