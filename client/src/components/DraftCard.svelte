<script lang="ts">
    import type { DraftItem } from "../lib/types";
    import { cardClass, cardIcon, cardTitle, cardType } from "../lib/cardInfo";

    export let item: DraftItem;
    export let disabled = false;
    export let selected = false;
</script>

<div
    class={[
        "market-card h-32 rounded-2xl border-2 p-3 text-left text-[#142833] shadow-[0_6px_0_rgba(0,0,0,0.18)] transition",
        selected ? "ring-4 ring-[#f2c36b]" : "",
        disabled ? "opacity-55" : "",
        cardClass(item),
    ].join(" ")}
>
    <div class="flex h-full flex-col justify-between">
        <div class="flex items-start justify-between gap-2">
            <div
                class="grid h-10 w-10 place-items-center rounded-xl bg-white/35 text-2xl font-black"
            >
                {cardIcon(item)}
            </div>

            <div
                class="rounded-lg bg-white/30 px-2 py-1 text-[10px] font-black uppercase tracking-wider"
            >
                {cardType(item)}
            </div>
        </div>

        <div>
            <div class="text-lg font-black leading-5">
                {cardTitle(item)}
            </div>
        </div>
    </div>
</div>

<style>
    .market-card {
        animation: card-refill 220ms ease-out;
    }

    @keyframes card-refill {
        0% {
            transform: translateY(8px) scale(0.97);
            opacity: 0;
        }

        100% {
            transform: translateY(0) scale(1);
            opacity: 1;
        }
    }
</style>
