<script lang="ts">
    export let action:
        | "outpost"
        | "settlement"
        | "city"
        | "blockade"
        | "floodworks"
        | "flood"
        | "pass" = "outpost";

    export let hint = "";

    let visible = false;
    let x = 0;
    let y = 0;

    function title(action: string) {
        switch (action) {
            case "outpost":
                return "Outpost";
            case "settlement":
                return "Settlement";
            case "city":
                return "City";
            case "blockade":
                return "Blockade";
            case "floodworks":
                return "Floodworks";
            case "flood":
                return "Flood";
            default:
                return "Build";
        }
    }

    function typeLabel(action: string) {
        switch (action) {
            case "city":
                return "Upgrade";
            case "blockade":
                return "Disruption";
            case "floodworks":
            case "flood":
                return "Relic";
            default:
                return "Build";
        }
    }

    function description(action: string) {
        switch (action) {
            case "outpost":
                return "Basic control structure. Build on empty non-river land. Used to claim or contest territory.";

            case "settlement":
                return "Production structure. Build only on friendly empty non-river land. Helps stabilize your economy.";

            case "city":
                return "Upgrade your own Outpost into a City. Cities are stronger and score better, but require an existing Outpost.";

            case "blockade":
                return "Disruptive structure. Place on enemy land, enemy structures, or adjacent neutral land. Cannot be placed on your own land, rivers, or already blockaded tiles.";

            case "floodworks":
                return "Spend Relic to buy flood tokens. Flood tokens let you convert valid land tiles into rivers.";

            case "flood":
                return "Use one flood token to convert a valid tile without a structure into a River.";

            default:
                return "";
        }
    }

    function clampPosition(nextX: number, nextY: number) {
        const tooltipWidth = 300;
        const tooltipHeight = 160;
        const padding = 16;

        x = Math.min(nextX, window.innerWidth - tooltipWidth - padding);
        y = Math.min(nextY, window.innerHeight - tooltipHeight - padding);

        x = Math.max(padding, x);
        y = Math.max(padding, y);
    }

    function showFromMouse(event: MouseEvent) {
        visible = true;
        clampPosition(event.clientX + 14, event.clientY + 14);
    }

    function moveFromMouse(event: MouseEvent) {
        if (!visible) return;
        clampPosition(event.clientX + 14, event.clientY + 14);
    }

    function showFromFocus(event: FocusEvent) {
        const target = event.currentTarget as HTMLElement;
        const rect = target.getBoundingClientRect();

        visible = true;
        clampPosition(rect.right + 14, rect.top);
    }

    function hide() {
        visible = false;
    }
</script>

<div
    class="block h-full"
    role="group"
    aria-label={`${title(action)} build details`}
    on:mouseenter={showFromMouse}
    on:mousemove={moveFromMouse}
    on:mouseleave={hide}
    on:focusin={showFromFocus}
    on:focusout={hide}
>
    <slot />
</div>

{#if visible}
    <div
        class="pointer-events-none fixed z-[2147483647] w-[300px] rounded-2xl bg-[#142833] p-4 text-[#f8efe0] shadow-2xl ring-1 ring-white/10"
        style={`left: ${x}px; top: ${y}px;`}
    >
        <div class="flex items-center justify-between gap-3">
            <div class="text-sm font-black">
                {title(action)}
            </div>

            <div
                class="rounded-lg bg-[#f8efe0]/10 px-2 py-1 text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
            >
                {typeLabel(action)}
            </div>
        </div>

        <div class="mt-2 text-xs font-semibold leading-5 text-[#d9e6df]">
            {description(action)}
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
