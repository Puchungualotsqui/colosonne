<script lang="ts">
    import type { DraftItem } from "../lib/types";
    import { cardDescription, cardTitle, cardType } from "../lib/cardInfo";

    export let item: DraftItem | null | undefined;
    export let hint = "";

    let visible = false;
    let x = 0;
    let y = 0;

    function clampPosition(nextX: number, nextY: number) {
        const tooltipWidth = 280;
        const tooltipHeight = 150;
        const padding = 16;

        x = Math.min(nextX, window.innerWidth - tooltipWidth - padding);
        y = Math.min(nextY, window.innerHeight - tooltipHeight - padding);

        x = Math.max(padding, x);
        y = Math.max(padding, y);
    }

    function showFromMouse(event: MouseEvent) {
        if (!item) return;

        visible = true;
        clampPosition(event.clientX + 14, event.clientY + 14);
    }

    function moveFromMouse(event: MouseEvent) {
        if (!visible || !item) return;

        clampPosition(event.clientX + 14, event.clientY + 14);
    }

    function showFromFocus(event: FocusEvent) {
        if (!item) return;

        if (visible) return;

        const target = event.target as HTMLElement | null;
        if (!target) return;

        const rect = target.getBoundingClientRect();

        visible = true;
        clampPosition(rect.right + 14, rect.top);
    }

    function hide() {
        visible = false;
    }
</script>

<div
    class="contents"
    role="group"
    aria-label={item ? `${cardTitle(item)} card details` : "Card details"}
    on:mouseenter={showFromMouse}
    on:mousemove={moveFromMouse}
    on:mouseleave={hide}
    on:focusin={showFromFocus}
    on:focusout={hide}
>
    <slot />
</div>

{#if visible && item}
    <div
        class="pointer-events-none fixed z-[2147483647] w-[280px] rounded-2xl bg-[#142833] p-4 text-[#f8efe0] shadow-2xl ring-1 ring-white/10"
        style={`left: ${x}px; top: ${y}px;`}
    >
        <div class="flex items-center justify-between gap-3">
            <div class="text-sm font-black">
                {cardTitle(item)}
            </div>

            <div
                class="rounded-lg bg-[#f8efe0]/10 px-2 py-1 text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
            >
                {cardType(item)}
            </div>
        </div>

        <div class="mt-2 text-xs font-semibold leading-5 text-[#d9e6df]">
            {cardDescription(item)}
        </div>

        {#if hint}
            <div
                class="mt-3 text-[10px] font-black uppercase tracking-[0.18em] text-[#f2c36b]"
            >
                {hint}
            </div>
        {/if}
    </div>
{/if}
