<script lang="ts">
    import DraftCard from "./DraftCard.svelte";
    import { GamePhase, type DraftItem, type GameState } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";

    export let onPick: (marketIndex: number) => void;

    const MARKET_SIZE = 6;
    const MAX_HAND = 3;

    $: me = game.Players.find((p) => p.Id === playerId);
    $: hand = me?.Hand ?? [];
    $: handCount = hand.length;

    $: marketSlots = Array.from(
        { length: MARKET_SIZE },
        (_, index) => game.Market[index] ?? null,
    );

    $: filledCount = marketSlots.filter(Boolean).length;

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId &&
        handCount < MAX_HAND;

    function pick(index: number, item: DraftItem) {
        debugLog("market.pick.click", {
            index,
            canPick,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            handCount,
            item,
        });

        if (!canPick) return;
        onPick(index);
    }
</script>

<section
    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
>
    <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-black text-[#fff7e8]">Market</h2>

        <div
            class={[
                "rounded-xl px-3 py-1 text-xs font-black uppercase tracking-wider",
                canPick
                    ? "bg-[#f2c36b] text-[#142833]"
                    : "bg-[#f8efe0]/10 text-[#9fc9c5]",
            ].join(" ")}
        >
            {#if canPick}
                {handCount}/3 hand
            {:else}
                {filledCount}/6 cards
            {/if}
        </div>
    </div>

    <div class="mt-4 grid grid-cols-2 gap-3">
        {#each marketSlots as item, index (index)}
            {#if item}
                <DraftCard
                    {item}
                    {index}
                    disabled={!canPick}
                    onPick={(pickedIndex) => pick(pickedIndex, item)}
                />
            {:else}
                <div
                    class="market-empty-slot grid h-32 place-items-center rounded-2xl border-2 border-dashed border-[#f8efe0]/20 bg-[#f8efe0]/6 p-3 text-center shadow-[0_6px_0_rgba(0,0,0,0.12)]"
                >
                    <div>
                        <div
                            class="mx-auto grid h-10 w-10 place-items-center rounded-xl bg-[#f8efe0]/10 text-2xl font-black text-[#9fc9c5]"
                        >
                            —
                        </div>

                        <div
                            class="mt-3 text-xs font-black uppercase tracking-[0.2em] text-[#9fc9c5]"
                        >
                            Empty
                        </div>

                        <div
                            class="mt-1 text-[10px] font-bold text-[#9fc9c5]/70"
                        >
                            Refills next pick
                        </div>
                    </div>
                </div>
            {/if}
        {/each}
    </div>
</section>

<style>
    .market-empty-slot {
        animation: empty-slot-in 180ms ease-out;
    }

    @keyframes empty-slot-in {
        0% {
            transform: scale(0.96);
            opacity: 0;
        }

        100% {
            transform: scale(1);
            opacity: 1;
        }
    }
</style>
