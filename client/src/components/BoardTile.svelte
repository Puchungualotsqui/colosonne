<script lang="ts">
    import { createEventDispatcher, onDestroy } from "svelte";
    import StructureGlyph from "./StructureGlyph.svelte";
    import { Structure, type Tile } from "../lib/types";
    import TileTooltip, {
        type InfluenceTooltipRow,
    } from "./TileTooltip.svelte";

    export let onTileHover: () => void = () => {};
    export let onTileLeave: () => void = () => {};

    export let left = 0;
    export let top = 0;
    export let width = 112;
    export let height = 98;

    export let tile: Tile | undefined = undefined;
    export let candidate = false;
    export let ghost = false;
    export let clickable = false;
    export let dimmed = false;

    export let biomeClass = "";
    export let structureLabel = "";
    export let tooltip = "";

    export let tooltipTitle = "";
    export let tooltipSubtitle = "";
    export let influenceRows: InfluenceTooltipRow[] = [];

    export let auraOwner = 0;
    export let auraEdges: boolean[] = [
        false,
        false,
        false,
        false,
        false,
        false,
    ];

    export let influencePreviewed = false;
    export let influencePreviewOwner = 0;

    export let onStructureHover: () => void = () => {};
    export let onStructureLeave: () => void = () => {};

    export let effect = "";
    export let effectOwner = 0;

    const dispatch = createEventDispatcher<{
        select: void;
    }>();

    let tooltipVisible = false;
    let tooltipX = 0;
    let tooltipY = 0;

    const edgeLines = [
        { x1: 57, y1: 8, x2: 97, y2: 29 },
        { x1: 97, y1: 29, x2: 97, y2: 69 },
        { x1: 97, y1: 69, x2: 57, y2: 90 },
        { x1: 55, y1: 90, x2: 15, y2: 69 },
        { x1: 15, y1: 69, x2: 15, y2: 29 },
        { x1: 15, y1: 29, x2: 55, y2: 8 },
    ];

    function auraClass(owner: number) {
        if (owner === 1) return "aura-p1";
        if (owner === 2) return "aura-p2";
        return "aura-tied";
    }

    function previewClass(owner: number) {
        if (owner === 1) return "preview-p1";
        if (owner === 2) return "preview-p2";
        return "preview-neutral";
    }

    function clampPosition(nextX: number, nextY: number) {
        const tooltipWidth = 280;
        const tooltipHeight = 220;
        const padding = 16;

        tooltipX = Math.min(nextX, window.innerWidth - tooltipWidth - padding);
        tooltipY = Math.min(
            nextY,
            window.innerHeight - tooltipHeight - padding,
        );

        tooltipX = Math.max(padding, tooltipX);
        tooltipY = Math.max(padding, tooltipY);
    }

    function showFromMouse(event: MouseEvent) {
        onTileHover();

        if (!tooltip && influenceRows.length === 0) return;

        tooltipVisible = true;
        clampPosition(event.clientX + 14, event.clientY + 14);

        if (tile && tile.Structure !== Structure.None) {
            onStructureHover();
        }
    }

    function moveFromMouse(event: MouseEvent) {
        if (!tooltipVisible) return;
        clampPosition(event.clientX + 14, event.clientY + 14);
    }

    function hide() {
        tooltipVisible = false;
        onStructureLeave();
        onTileLeave();
    }

    function structureOwnerClass(owner: number) {
        if (owner === 1) {
            return "bg-[#1d4e89] text-white ring-[#f8efe0]/35";
        }

        if (owner === 2) {
            return "bg-[#b94b3f] text-white ring-[#f8efe0]/35";
        }

        return "bg-[#f8efe0]/70 text-[#142833] ring-black/10";
    }

    function portal(node: HTMLElement) {
        document.body.appendChild(node);

        return {
            destroy() {
                node.remove();
            },
        };
    }

    function selectTile(event: MouseEvent) {
        event.preventDefault();
        event.stopPropagation();

        // Do not block here.
        // Board.svelte already has the authoritative isClickable(hex) check.
        // Blocking here makes manual build clicks fail when the visual clickable prop
        // is briefly stale or when selectedBuildAction changes in the same frame.
        dispatch("select");
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            dispatch("select");
        }
    }

    onDestroy(() => {
        onStructureLeave();
    });

    let rootEl: HTMLDivElement;

    export function getCenterViewport() {
        if (!rootEl) return null;

        const rect = rootEl.getBoundingClientRect();

        return {
            x: rect.left + rect.width / 2,
            y: rect.top + rect.height / 2,
        };
    }
</script>

<div
    bind:this={rootEl}
    class={[
        "group board-tile clip-hex absolute box-border flex items-center justify-center border-[2px] shadow-[0_7px_0_rgba(74,48,31,0.22)] transition",
        biomeClass,
        clickable
            ? "cursor-pointer ring-4 ring-[#f2c36b] ring-offset-2 ring-offset-[#d9c291] brightness-110 saturate-110 hover:brightness-125"
            : "cursor-default",
        candidate ? "border-dashed opacity-85" : "",
        ghost ? "pointer-events-none opacity-0" : "",
        dimmed ? "tile-dimmed" : "",
        influencePreviewed
            ? `structure-preview ${previewClass(influencePreviewOwner)}`
            : "",
        effect === "reinforce"
            ? `tile-reinforce tile-reinforce-p${effectOwner}`
            : "",
    ].join(" ")}
    style={`left: ${left}px; top: ${top}px; width: ${width}px; height: ${height}px;`}
    role="button"
    tabindex={clickable ? 0 : -1}
    aria-label={tooltip}
    aria-disabled={!clickable}
    on:mouseenter={showFromMouse}
    on:mousemove={moveFromMouse}
    on:mouseleave={hide}
    on:click={selectTile}
    on:keydown={handleKeydown}
>
    <div
        class="pointer-events-none absolute inset-[4px] z-[1] clip-hex border border-white/18"
    ></div>

    {#if auraOwner > 0}
        <svg
            class="pointer-events-none absolute inset-0 z-[2] h-full w-full"
            viewBox="0 0 112 98"
            preserveAspectRatio="none"
            aria-hidden="true"
        >
            {#each edgeLines as edge, index}
                {#if auraEdges[index]}
                    <line
                        class={["aura-edge", auraClass(auraOwner)].join(" ")}
                        x1={edge.x1}
                        y1={edge.y1}
                        x2={edge.x2}
                        y2={edge.y2}
                    />
                {/if}
            {/each}
        </svg>
    {/if}

    {#if candidate}
        <div
            class="relative z-10 grid h-10 w-10 place-items-center rounded-full bg-[#f8efe0]/70 text-2xl font-black text-[#6b4a2f]"
        >
            +
        </div>
    {:else if tile}
        {#if tile.Structure !== Structure.None}
            <div class="relative z-20">
                <StructureGlyph
                    structure={tile.Structure}
                    owner={tile.StructureOwner === 1
                        ? 1
                        : tile.StructureOwner === 2
                          ? 2
                          : "neutral"}
                    size="md"
                    boxed
                    title={`${structureLabel} · P${tile.StructureOwner}`}
                />
            </div>
        {/if}

        {#if tile.HasBlockade}
            <div
                class={[
                    "absolute right-3 top-3 z-20 grid h-6 w-6 place-items-center rounded-full text-xs font-black shadow-sm",
                    tile.BlockadeOwner === 1
                        ? "bg-[#1d4e89] text-white"
                        : tile.BlockadeOwner === 2
                          ? "bg-[#b94b3f] text-white"
                          : "bg-[#142833] text-white",
                ].join(" ")}
                title={`Blockade P${tile.BlockadeOwner}`}
            >
                ✕
            </div>
        {/if}
    {/if}
</div>

{#if tooltipVisible}
    <div
        use:portal
        class="pointer-events-none fixed w-[280px]"
        style={`left: ${tooltipX}px; top: ${tooltipY}px; z-index: 2147483647;`}
    >
        <TileTooltip
            title={tooltipTitle || tooltip}
            subtitle={tooltipSubtitle}
            {influenceRows}
        />
    </div>
{/if}

<style>
    .clip-hex {
        clip-path: polygon(
            50% 0%,
            93.3% 25%,
            93.3% 75%,
            50% 100%,
            6.7% 75%,
            6.7% 25%
        );
    }

    .board-tile {
        box-sizing: border-box;
        overflow: hidden;
        min-width: 0;
        min-height: 0;
        line-height: 1;
        transform: none;
        isolation: isolate;
    }

    .board-tile * {
        box-sizing: border-box;
    }

    .board-tile.tile-dimmed {
        filter: saturate(0.88) brightness(0.96);
        opacity: 0.82;
    }

    .board-tile.tile-dimmed::before {
        content: "";
        position: absolute;
        inset: 0;
        z-index: 9;
        pointer-events: none;
        background:
            linear-gradient(
                rgba(248, 239, 224, 0.12),
                rgba(248, 239, 224, 0.12)
            ),
            repeating-linear-gradient(
                -35deg,
                rgba(20, 40, 51, 0.07) 0px,
                rgba(20, 40, 51, 0.07) 2px,
                transparent 2px,
                transparent 10px
            );
    }

    .aura-edge {
        stroke-width: 8;
        stroke-linecap: round;
        stroke-linejoin: round;
        opacity: 0.92;
        vector-effect: non-scaling-stroke;
    }

    .aura-p1 {
        stroke: #1d4e89;
    }

    .aura-p2 {
        stroke: #b94b3f;
    }

    .aura-tied {
        stroke: #f8efe0;
        opacity: 0.62;
    }

    .structure-preview {
        z-index: 8;
    }

    .structure-preview::after {
        content: "";
        position: absolute;
        inset: 7px;
        z-index: 3;
        clip-path: polygon(
            50% 0%,
            93.3% 25%,
            93.3% 75%,
            50% 100%,
            6.7% 75%,
            6.7% 25%
        );
        pointer-events: none;
        background: rgba(248, 239, 224, 0.22);
        border: 4px solid rgba(248, 239, 224, 0.92);
        filter: drop-shadow(0 0 8px rgba(248, 239, 224, 0.32));
    }

    .preview-p1::after {
        background: rgba(29, 78, 137, 0.2);
        border-color: rgba(29, 78, 137, 0.95);
        filter: drop-shadow(0 0 7px rgba(29, 78, 137, 0.32));
    }

    .preview-p2::after {
        background: rgba(185, 75, 63, 0.2);
        border-color: rgba(185, 75, 63, 0.95);
        filter: drop-shadow(0 0 7px rgba(185, 75, 63, 0.32));
    }

    .board-tile.tile-reinforce::after {
        content: "";
        position: absolute;
        inset: 7px;
        z-index: 30;
        pointer-events: none;
        clip-path: polygon(
            50% 0%,
            93.3% 25%,
            93.3% 75%,
            50% 100%,
            6.7% 75%,
            6.7% 25%
        );
        border: 5px solid rgba(248, 239, 224, 0.95);
        background: rgba(248, 239, 224, 0.16);
        animation: reinforce-pulse 1250ms ease-out forwards;
    }

    .board-tile.tile-reinforce-p1::after {
        border-color: rgba(29, 78, 137, 0.95);
        background: rgba(29, 78, 137, 0.18);
    }

    .board-tile.tile-reinforce-p2::after {
        border-color: rgba(185, 75, 63, 0.95);
        background: rgba(185, 75, 63, 0.18);
    }

    @keyframes reinforce-pulse {
        0% {
            opacity: 0;
            transform: scale(0.72);
            filter: drop-shadow(0 0 0 rgba(248, 239, 224, 0));
        }

        18% {
            opacity: 1;
            transform: scale(1.02);
            filter: drop-shadow(0 0 12px rgba(248, 239, 224, 0.55));
        }

        58% {
            opacity: 0.92;
            transform: scale(1.14);
        }

        100% {
            opacity: 0;
            transform: scale(1.32);
            filter: drop-shadow(0 0 0 rgba(248, 239, 224, 0));
        }
    }
</style>
